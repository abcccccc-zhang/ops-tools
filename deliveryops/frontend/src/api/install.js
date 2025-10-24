import axios from "axios";

export function getInstallPackageList(version, type, arch) {
  return axios.post("/api/getpackagelist", {
    version,
    type,
    arch,
  });
}

export function generateDownloadURL(path, expire) {
  return axios.post("/api/generateDownloadURL", {
    path,
    expire,
  });
}