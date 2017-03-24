package main

import (
        "log"
        "os"

        "github.com/danehans/gonx/nx"
)

func main() {

        if len(os.Args) < 3 {
                log.Fatalf("usage: %s add|remove|replace|show [ethernet|port-channel][:id] [allowed-vlan] [native-vlan] ",
                           os.Args[0])
                log.Fatalf("       set native-vlan to None to remove existing config.")
        }

        cmd := os.Args[1]
        which := os.Args[2]
        var allowed, native string
        if len(os.Args) > 3 {
                allowed = os.Args[3]
        }
        if len(os.Args) > 4 {
                native = os.Args[4]
        }

        a := login(true)
        defer logout(a)

        // add/del/replace ethernet interface
        execute(a, cmd, which, allowed, native)

        resp, getErr := a.GetInterface(which)
        if getErr != nil {
                log.Printf("could not get/list interface: %v", getErr)
                return
        }

        // Print the legend
        log.Printf("ID\t\tNative\tTrunk\tMode\tState\tDescr\n")
        for _, r := range resp {
                tvlan := r["trunkVlans"]
                if r["trunkVlans"] == "" {
                    tvlan = "None"
                }
                log.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", r["id"], r["nativeVlan"],
                           tvlan, r["mode"], r["adminSt"], r["descr"])
        }

}

func execute(a *nx.Client, cmd string, ifName string,
allowed string, native string) {
        switch cmd {
        case "add":
                allowed = "+" + allowed
        case "replace":
        case "remove":
                allowed = "-" + allowed
                if native != "" {
                    native = "None"
                }
        case "show":
                return
        default:
                log.Printf("unknown command: %s", cmd)
                return
        }

 
        // Note client package uses add naming instead of create.
        err := a.AddTrunkVlan(ifName,
                              allowed, native)
        if err != nil {
                os.Exit(1)
        }

        if native != "" {
            log.Printf("Trunk Vlan %s native %s %s for interface %s\n",
                       allowed, native, cmd, ifName)
        } else {
            log.Printf("Trunk Vlan %s %s for interface.\n",
                       allowed, cmd, ifName)
        }
}

func login(debug bool) *nx.Client {

        a, errNew := nx.New(nx.ClientOptions{Debug: debug})
        if errNew != nil {
                log.Printf("login new client error: %v", errNew)
                os.Exit(1)
        }
        errLogin := a.Login()
        if errLogin != nil {
                log.Printf("login error: %v", errLogin)
                os.Exit(1)
        }

        return a
}

func logout(a *nx.Client) {
        a.Logout()

        log.Printf("logout: done")
}

