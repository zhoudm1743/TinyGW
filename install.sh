#!/bin/bash
set -euo pipefail

# 必须root运行
[[ $EUID -eq 0 ]] || { echo "错误：请以root用户执行此脚本" >&2; exit 1; }

# 安装目录配置
INSTALL_DIR="/www/wwwroot/tinyGW"
mkdir -p "$INSTALL_DIR"

# 依赖安装
install_deps() {
    source /etc/os-release 2>/dev/null || source /usr/lib/os-release 2>/dev/null || { echo "错误：无法确定发行版"; exit 1; }

    case "$ID" in
        debian|ubuntu|deepin)
            echo "安装依赖（Debian系）..."
            apt update -y
            apt install -y wget lua5.3 unzip
            ;;
        centos|fedora|almalinux|rocky|opencloudos|alinux)
            echo "安装依赖（RHEL系）..."
            if ! rpm -q epel-release &>/dev/null; then
                yum install -y epel-release || { echo "安装EPEL仓库失败"; exit 1; }
            fi
            if command -v yum &>/dev/null; then
                yum install -y wget lua53 unzip
            elif command -v dnf &>/dev/null; then
                dnf install -y wget lua53 unzip
            else
                echo "错误：未找到yum或dnf"; exit 1
            fi
            ;;
        *)
            echo "不支持的发行版：$ID"; exit 1
            ;;
    esac
}

install_deps

echo "正在下载文件..."
wget http://utils.zsxiao.cn/gateway/tinyGW.zip -O "$INSTALL_DIR/tinyGW.zip"

echo "正在解压文件..."
unzip -qo "$INSTALL_DIR/tinyGW.zip" -d "$INSTALL_DIR" || { echo "解压失败" >&2; exit 1; }

# 配置管理模块
CONFIG_DIR="$INSTALL_DIR/config"
mkdir -p "$CONFIG_DIR"
cp "$(dirname "$0")/config/conf.yml" "$CONFIG_DIR/" || { echo "配置文件复制失败" >&2; exit 1; }

# 交互式配置参数
read -p "请输入网关ClientID（默认：zsxagwA001）: " gw_clientid
read -p "请输入订阅服务器IP（默认：mqtt.zsxakj.com）: " sub_ip
read -p "请输入订阅用户名（默认：bgscs）: " sub_user
read -p "请输入订阅密码（默认：123456）: " -s sub_pass
echo
read -p "请输入订阅端口（默认：1883）: " sub_port

# 设置默认值
gw_clientid=${gw_clientid:-zsxagwA001}
sub_ip=${sub_ip:-mqtt.zsxakj.com}
sub_user=${sub_user:-bgscs}
sub_pass=${sub_pass:-123456}
sub_port=${sub_port:-1883}

# 更新配置文件
sed -i \
-e "/^gateway:/,/^$/ {/^    clientid:/ s/.*/    clientid: $gw_clientid/}" \
-e "/^subscribe:/,/^$/ {/^    ip:/ s/.*/    ip: $sub_ip/}" \
-e "/^subscribe:/,/^$/ {/^    port:/ s/.*/    port: \"$sub_port\"/}" \
-e "/^subscribe:/,/^$/ {/^    username:/ s/.*/    username: $sub_user/}" \
-e "/^subscribe:/,/^$/ {/^    password:/ s/.*/    password: \"$sub_pass\"/}" \
"$CONFIG_DIR/conf.yml"

# 设置权限
chown -R root:root "$INSTALL_DIR"
chmod 755 "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/tinyGW"

# 处理服务文件
SERVICE_FILE="$INSTALL_DIR/tinygw.service"
if [[ -f "$SERVICE_FILE" ]]; then
    mv "$SERVICE_FILE" /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable --now tinygw || { echo "服务启动失败" >&2; exit 1; }
else
    echo "警告：未找到服务文件，请手动配置" >&2
fi

echo "安装成功！正在启动"

# 获取本机IP地址（兼容多系统）
get_local_ip() {
    local ipv4_address=""
    # 尝试通过多种方式获取IP
    ipv4_address=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -z "$ipv4_address" ] && ipv4_address=$(ip route get 1 2>/dev/null | awk '{print $(NF-2);exit}')
    [ -z "$ipv4_address" ] && ipv4_address=$(curl -s icanhazip.com 2>/dev/null)
    echo "${ipv4_address:-127.0.0.1}"
}

ACCESS_IP=$(get_local_ip)
echo "------------------------------"
echo "访问面板的连接地址："
echo -e "内网访问：\033[32mhttp://${ACCESS_IP}:1883\033[0m"
echo -e "公网访问：\033[33mhttp://$(curl -s icanhazip.com 2>/dev/null || echo 公网IP):1883\033[0m"
echo "------------------------------"
echo "防火墙提示："
echo "若无法访问，请确保已开放端口："
echo "sudo firewall-cmd --permanent --add-port=1883/tcp && firewall-cmd --reload  # CentOS"
echo "sudo ufw allow 1883/tcp  # Ubuntu/Debian"