


# veth

namespace与namespace间的交互

`测试`
```shell script
# 创建2个网络namespace
ip netns add ns1
ip netns add ns2
# 创建一对veth
ip link add veth0 type veth peer name veth1
# 移动veth到namespace
ip link set veth0 netns ns1
ip link set veth1 netns ns2
# 查看namespace下的网络设备
ip netns exec ns1 ip link
ip netns exec ns2 ip link
# 设置ip地址
ip netns exec ns1 ip addr add  172.18.0.2/24  dev veth0
ip netns exec ns2 ip addr add  172.18.0.3/24  dev veth1
# 启动veth
ip netns exec ns1 ip link set veth0 up
ip netns exec ns2 ip link set veth1 up
# 查看veth
ip netns exec ns1 ip a
ip netns exec ns2 ip a
# ping测试
ip netns exec ns1 ping 172.18.0.2
ip netns exec ns1 ping 172.18.0.3
ip netns exec ns1 ping 172.18.0.2
ip netns exec ns2 ping 172.18.0.3
```

# bridge

namespace与物理机间的交互

`依赖包`
```
yum -y install tunctl bridge-utils
```
`测试`

```shell script
# 创建网络namespace
ip netns add nsx
# 创建一对veth
ip link add vethx1 type veth peer name vethx2
# 移动veth到namespace
ip link set vethx2 netns nsx
# 查看namespace下的网络设备
ip netns exec ns4 ip link
# 添加网卡到网桥
brctl addif brx vethx1
brctl addif brx enp0s3
# 启动网桥和vethx1
ip link set vethx1 up
ip link set brx up
# 查看nsx的虚拟网卡
ip netns exec nsx ip link
# 添加namespace nsx内虚拟网卡ip
ip netns exec  nsx ip addr add  172.18.0.11/24 dev vethx2
# 启动namespace nsx内虚拟网卡vethx2
ip netns exec nsx ip link set vethx2 up
# namespace nsx内虚拟网卡添加路由
 ip netns exec nsx ip route add default dev vethx2
# 网桥添加路由
ip route add  172.18.0.0/24 dev brx
# ping测试
ip addr 
ping 172.18.0.11
ip netns exec nsx ping 172.18.0.11
ip netns exec nsx ping 192.168.2.199
ip netns exec nsx ping 192.168.56.102
```

# 流程
1、创建网络

mydocker network create --subnet 192.168.0.0/24 --driver bridge testbridgenet

2、创建容器链接网络

mydocker run -ti -p 80 : 80 --net testbridgenet xxxx
