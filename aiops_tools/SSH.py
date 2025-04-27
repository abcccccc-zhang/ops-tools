from langchain_core.output_parsers import JsonOutputParser
from pydantic import BaseModel, Field
from langchain_openai import ChatOpenAI
import os
import  paramiko

# os.environ["OPENAI_API_BASE"] = 'https://'
os.environ["OPENAI_API_BASE"] = 'https://'
# os.environ["OPENAI_API_KEY"] = ''
from langchain_core.prompts import PromptTemplate
def execute_ssh_command(host="", port=22, username="", password="", command=""):
    try:
        # 创建 SSH 客户端对象
        client = paramiko.SSHClient()
        # 自动添加主机密钥
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        # 连接到远程主机
        client.connect(host, port=port, username=username, password=password)
        # 设置环境变量（例如设置 TERM）
        env = {
            'TERM': 'xterm'
        }
        # 执行命令
        stdin, stdout, stderr = client.exec_command(command, environment=env)

        # 获取命令输出
        output = stdout.read().decode()
        error = stderr.read().decode()

        # 关闭连接
        client.close()

        if output:
            return output
        elif error:
            return error
        else:
            return "命令执行成功，但没有输出。"
    except Exception as e:
        return f"SSH 连接或命令执行失败: {str(e)}"

class SSHInput(BaseModel):
    host: str = Field(description="SSH 主机地址")
    port: int = Field(default=22, description="SSH 端口")
    username: str = Field(description="SSH 用户名")
    password: str = Field(description="SSH 密码")
    command: str = Field(description="需要执行的命令")
    directory: str = Field(description="目标目录")
# 初始化 OpenAI 模型
chat = ChatOpenAI(model_name="gpt-4o-mini",temperature=0)
# chat = ChatOpenAI(model_name="deepseek-reasoner",temperature=0)

# 设置解析器，注入模板
parser = JsonOutputParser(pydantic_object=SSHInput)
# {format_instructions}
# 修改提示模板以便更好理解并清晰指引模型处理
prompt = PromptTemplate(
    template="""
    请根据以下信息解析，并生成一个 SSH 请求。
    {format_instructions}
    用户查询: {query}
    输出: 需要执行的有效命令 如果用户没有给出命令，则需要你帮助
    """,
    input_variables=["query"],
    partial_variables={"format_instructions": parser.get_format_instructions()}
)
# print(parser.get_format_instructions())
# 创建链条
chain = prompt | chat | parser

# 示例查询，注意这里要给定足够的具体信息  top -b -i -n 1
query = " 远程192.168.12.4 账号root 密码123456 查看系统CPU负载 "
# 发送消息并获取响应
response = chain.invoke({"query": query})
# 输出解析结果
print(response)
#
# 使用 paramiko 执行 SSH 命令
ssh_input = response  # 假设 response 结果是解析后的 SSHInput 对象
port = ssh_input.get('port', 22)

output = execute_ssh_command(
    host=ssh_input['host'],
    port=port,  # 使用处理后的 port
    username=ssh_input['username'],
    password=ssh_input['password'],
    command=ssh_input['command']
)

# 输出远程执行结果
print(output)
