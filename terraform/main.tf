terraform {
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = "3.0.1-rc4"
    }
    macaddress = {
      source = "ivoronin/macaddress"
      version = "0.3.0"
    }
  }
}
provider "proxmox" {
  pm_tls_insecure     = true
  pm_api_url          = "https://192.168.1.200:8006/api2/json"
  pm_api_token_id     = "terraform-prov@pve!terraform-token"
  pm_api_token_secret = "f23ff08e-a54e-4adb-bd32-c70a7f662d34"
}
variable "master_vm_id" {
  type = list(number)
  default = [170,168,171]
}

variable "agent_vm_id" {
  type = list(number)
  default = [169]
}

variable "master_ip" {
  type    = list(string)
  default = ["192.168.1.167", "192.168.1.168","192.168.1.170"]
}

variable "agent_ips" {
  type    = list(string)
  default = ["192.168.1.169"]
}
variable "master_vm_count" {
  type    = number
  default = 3
}

variable "agent_vm_count" {
  type    = number
  default = 1
}
# 初始化节点的 Cloud-init 配置
data "template_file" "master_init_user_data" {
  template = file("${path.module}/cloud-init/master-init-user-data.tmpl")
  vars = {
    pubkey   = file("${path.module}/cloud-init/home_id_ed25519.pub")
    hostname = "k3s-master-1"
    passwd   = "123456"
    token    = "123456"  # K3S 集群初始化令牌
  }
}

# 加入集群的主节点 Cloud-init 配置
data "template_file" "master_join_user_data" {
  count    = var.master_vm_count - 1
  template = file("${path.module}/cloud-init/master-join-user-data.tmpl")
  vars = {
    pubkey   = file("${path.module}/cloud-init/home_id_ed25519.pub")
    hostname = format("k3s-master-%d", count.index + 2)
    passwd   = "123456"
    token    = "123456"  # K3S 集群初始化令牌
    joinip   = var.master_ip[0]  # 第一个主节点的 IP
    joinport = "6443"
  }
}

# 工作节点的 Cloud-init 配置

data "template_file" "agent_user_data" {
  template = file("${path.module}/cloud-init/agent-user-data.tmpl")
  vars = {
    pubkey   = file("${path.module}/cloud-init/home_id_ed25519.pub")
    hostname = "k3s-agent-1"
    passwd   = "123456"
    token    = "123456"  # K3S 集群初始化令牌
    joinip   = var.master_ip[0]
    joinport = "6443"
  }
}

resource "local_file" "master_init_user_data" {
  content  = data.template_file.master_init_user_data.rendered
  filename = "${path.module}/files/master_init_user_data.cfg"
}

resource "local_file" "master_join_user_data" {
  count    = length(data.template_file.master_join_user_data)
  content  = data.template_file.master_join_user_data[count.index].rendered
  filename = "${path.module}/files/master_join_user_data-${count.index}.cfg"
}

resource "local_file" "agent_user_data" {
  content  = data.template_file.agent_user_data.rendered
  filename = "${path.module}/files/agent_user_data.cfg"
}
resource "null_resource" "cloud_init_upload" {
  connection {
    type     = "ssh"
    user     = "root"
    password = "Zhang.123"
    host     = "192.168.1.200"
  }
  count = length(data.template_file.master_join_user_data)

  provisioner "file" {
    source      = local_file.master_init_user_data.filename
    destination = "/var/lib/vz/snippets/master_init_user_data.yml"
  }

   provisioner "file" {
    source      = local_file.master_join_user_data[count.index].filename
    destination = "/var/lib/vz/snippets/master_join_user_data_${count.index + 1}.yml"
  }

  provisioner "file" {
    source      = local_file.agent_user_data.filename
    destination = "/var/lib/vz/snippets/agent_user_data.yml"
  }
}
resource "proxmox_vm_qemu" "master" {
  count       = var.master_vm_count
  vmid        = var.master_vm_id[count.index]
  depends_on  = [null_resource.cloud_init_upload]
  name        = "k3s-master-${count.index + 1}"
  target_node = "zhang"
  clone       = "cloud-init-k8s-env"
  agent       = 1
  os_type     = "cloud-init"
  cores       = 2
  sockets     = 4
  cpu         = "host"
  memory      = 8192
  scsihw      = "virtio-scsi-single"
  bootdisk    = "scsi0"
  ipconfig0   = "ip=${var.master_ip[count.index]}/24,gw=192.168.1.1"
  disks {
    ide {
      ide0 {
        cloudinit {
          storage = "nas"
        }
      }
    }
    scsi {
      scsi0 {
        disk {
          size   = 32
          storage = "local-lvm"
        }
      }
    }
  }
   # cicustom    = "user=local:snippets/${count.index == 0 ? "master_init_user_data" : "master_join_user_data"}.yml"
  cicustom = "user=local:snippets/${count.index == 0 ? "master_init_user_data" : "master_join_user_data_${count.index}"}.yml"

}

# Agent nodes configuration
resource "proxmox_vm_qemu" "agent" {
  count       = var.agent_vm_count
  vmid        = var.agent_vm_id[count.index]
  depends_on  = [null_resource.cloud_init_upload, proxmox_vm_qemu.master]
  name        = "k3s-agent-${count.index + 1}"
  target_node = "zhang"
  clone       = "cloud-init-k8s-env"
  agent       = 1
  os_type     = "cloud-init"
  cores       = 2
  sockets     = 4
  cpu         = "host"
  memory      = 8192
  scsihw      = "virtio-scsi-single"
  bootdisk    = "scsi0"
  ipconfig0   = "ip=${var.agent_ips[count.index]}/24,gw=192.168.1.1"
  cicustom    = "user=local:snippets/agent_user_data.yml"
  disks {
    ide {
      ide0 {
        cloudinit {
          storage = "nas"
        }
      }
    }
    scsi {
      scsi0 {
        disk {
          size   = 32
          storage = "local-lvm"
        }
      }
    }
  }
}
