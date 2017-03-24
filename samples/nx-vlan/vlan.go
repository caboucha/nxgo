package main

import (
        "log"
        "os"

        "github.com/danehans/gonx/nx"
)

func main() {

        var vlanid, vni string

        if len(os.Args) < 2 {
                log.Fatalf("usage: %s add|remove|show [id] [segment]", os.Args[0])
        }
        cmd := os.Args[1]

        if len(os.Args) > 2 {
            vlanid = os.Args[2]
        }
        if len(os.Args) > 3 {
            vni = os.Args[3]
        }

        a := login(true)
        defer logout(a)

        // add/del vlans
        execute(a, cmd, vlanid, vni)

        resp, getErr := a.GetVlan(vlanid)
        if getErr != nil {
                log.Printf("could not get vlan: %v", getErr)
                return
        }

        // Print the legend
        log.Printf("\tID\t\tVNI\t\tAdmin State\tOperState\tName\n")
        for _, r := range resp {
                if r["BdOperName"] == "" {
                    continue
                }
                if r["accEncap"] == "unknown" {
                    log.Printf("\t%s\t\t%s\t\t%s\t\t%s\t\t%s\n", r["id"], r["accEncap"],
                               r["adminSt"], r["operSt"], r["BdOperName"])
                } else {
                    log.Printf("\t%s\t\t%s\t%s\t\t%s\t\t%s\n", r["id"], r["accEncap"],
                               r["adminSt"], r["operSt"], r["BdOperName"])
                }
        }

}

func execute(a *nx.Client, cmd string, vlanid string, vni string) {
        switch cmd {
        case "add":
            a.AddVlan(vlanid, vni)
        case "remove":
            a.DeleteVlan(vlanid)
        case "show":
                return
        default:
                log.Printf("unknown command: %s", cmd)
                return
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
