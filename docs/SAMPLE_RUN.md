This documentation explains how to run sample code which drivers API code.

First set-up your environment variables.
export NEXUS_HOSTS=ip-address-of-your-nexus-switch
export NEXUS_USER=administrator-user-name
export NEXUS_PASS=administrator-password
export NEXUS_DEBUG    # optional boolean to enable debug

1) Sample Ethernet interface usage
* go run samples/nx-interface/interface.go
2017/03/22 10:29:49 usage: /tmp/go-build075401116/command-line-arguments/_obj/exe/interface add|remove|replace|show [ethernet|port-channel][:id] [allowed-vlan] [native-vlan]

* go run samples/nx-interface/interface.go show ethernet
2017/03/16 15:05:32 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:05:32 eth1/33     vlan-1  1-4094  access  up
2017/03/16 15:05:32 eth1/34     vlan-1  1-4094  access  up
2017/03/16 15:05:32 eth1/35     vlan-1  1-4094  access  up
2017/03/16 15:05:32 eth1/36     vlan-1  1-4094  access  up
2017/03/16 15:05:32 eth1/37     vlan-1  1-4094  access  up
<SNIP>

* go run samples/nx-interface/interface.go show ethernet:1/19
2017/03/16 15:05:03 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:05:03 eth1/19     vlan-1  None    trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go add ethernet:1/19 197-199,200 197
2017/03/16 15:08:36 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:08:36 eth1/19     vlan-197        197-200 trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go replace ethernet:1/19 None None
2017/03/16 15:11:28 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:11:28 eth1/19     vlan-1  None    trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go add ethernet:1/19 197-199,200 197
Same as earlier then
go run samples/nx-interface/interface.go remove ethernet:1/19 197-199,200 197
2017/03/16 15:13:01 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:13:01 eth1/19     vlan-1  None    trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go replace ethernet:1/19 197-199,200 197
Same as earlier then
go run samples/nx-interface/interface.go replace ethernet:1/19 193-194 None
2017/03/16 15:14:16 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:14:16 eth1/19     vlan-1  193-194 trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go add ethernet:1/19 188 188
2017/03/16 15:14:57 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:14:57 eth1/19     vlan-188        188,193-194     trunk   up      connection to UCS bxb-ds-46

* go run samples/nx-interface/interface.go remove ethernet:1/19 188 188
2017/03/16 15:15:30 ID          Native  Trunk   Mode    State   Descr
2017/03/16 15:15:30 eth1/19     vlan-1  193-194 trunk   up      connection to UCS bxb-ds-46

2) Sample Port-channel interface tests
Do the same as above but use port-channel:55 instead of ethernet:1/19

3) Sample vlan usage
* go run samples/nx-vlan/vlan.go
2017/03/22 10:31:00 usage: /tmp/go-build224776088/command-line-arguments/_obj/exe/vlan add|remove|show [id] [segment]

* go run samples/nx-vlan/vlan.go show
2017/03/22 10:19:33     ID              VNI             Admin State     OperState       Name
2017/03/22 10:19:33     193             vxlan-70050     active          down            VLAN0193
2017/03/22 10:19:33     505             unknown         active          down            VLAN0505

* go run samples/nx-vlan/vlan.go add 223 70014
2017/03/22 10:21:10     ID              VNI             Admin State     OperState       Name
2017/03/22 10:21:10     223             vxlan-70014     active          down            VLAN0223

* go run samples/nx-vlan/vlan.go show
2017/03/22 10:21:49     ID              VNI             Admin State     OperState       Name
2017/03/22 10:21:49     193             vxlan-70050     active          down            VLAN0193
2017/03/22 10:21:49     505             unknown         active          down            VLAN0505
2017/03/22 10:21:49     223             vxlan-70014     active          down            VLAN0223

* go run samples/nx-vlan/vlan.go remove 223
go run samples/nx-vlan/vlan.go show
2017/03/22 10:22:33     ID              VNI             Admin State     OperState       Name
2017/03/22 10:22:33     193             vxlan-70050     active          down            VLAN0193
2017/03/22 10:22:33     505             unknown         active          down            VLAN0505

* go run samples/nx-vlan/vlan.go show 193
2017/03/22 10:23:03     ID              VNI             Admin State     OperState       Name
2017/03/22 10:23:03     193             vxlan-70050     active          down            VLAN0193

* go run samples/nx-vlan/vlan.go add 222
2017/03/22 10:33:16     ID              VNI             Admin State     OperState       Name
2017/03/22 10:33:16     222             unknown         active          down            VLAN0222

