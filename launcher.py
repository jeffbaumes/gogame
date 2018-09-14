import tkinter as tk
import os
root = tk.Tk()
worldd = "default"
portd = 5555
seedd = 1
named = "andrew"
cportd = 5555
hostd = "localhost"
# print('./buildorb {} {} {} {} {} {} {}'.format(world, port, seed, "1", "2", "3", "server"))
def ssumit():
    global worldd
    global portd
    global seedd
    global named
    global cportd
    global hostd
    worldst = world.get()
    portst = port.get()
    seedst = seed.get()
    if worldst != "":
        worldd = worldst
    if portst != "":
        portd = portst
    if seedst != "":
        seedd = seedst
    # print(worlds, ports, seeds)
    root.destroy()
    os.system('./buildorb {} {} {} {} {} {} {}'.format(worldd, portd, seedd, "1", "2", "3", "server"))
def csumit():
    global worldd
    global portd
    global seedd
    global named
    global cportd
    global hostd
    namest = name.get()
    hostst = host.get()
    cportst = cport.get()
    if worldst != "":
        named = namest
    if cportst != "":
        cportd = cportst
    if hostst != "":
        hostd = hostst
    # print(worlds, ports, seeds)
    root.destroy()
    os.system('./buildorb {} {} {} {} {} {} {}'.format("1", "1", "1", named, hostd, cportd, "server"))
def bsumit():
    global worldd
    global portd
    global seedd
    global named
    global cportd
    global hostd
    worldst = world.get()
    portst = port.get()
    seedst = seed.get()
    if worldst != "":
        worldd = worldst
    if portst != "":
        portd = portst
    if seedst != "":
        seedd = seedst
    namest = name.get()
    hostst = host.get()
    cportst = cport.get()
    if worldst != "":
        named = namest
    if cportst != "":
        cportd = cportst
    if hostst != "":
        hostd = hostst
    root.destroy()
    os.system('./buildorb {} {} {} {} {} {} {}'.format(worldd, portd, seedd, named, hostd, cportd, "all"))
def server():
    global world
    global port
    global seed
    global name
    global cport
    global host
    runs.destroy()
    runs1.destroy()
    runs2.destroy()
    world = tk.Entry(root)
    worldl = tk.Label(root, text="enter you world name leave blank for default" )
    port = tk.Entry(root)
    portl = tk.Label(root, text="enter you port for the server leave blank for 5555")
    seed = tk.Entry(root)
    seedl = tk.Label(root, text="enter the seed for the world")
    worldl.pack()
    world.pack()
    portl.pack()
    port.pack()
    seedl.pack()
    seed.pack()
    sumit = tk.Button(root, text="summit", command=ssumit)
    sumit.pack()
def client():
    global world
    global port
    global seed
    global name
    global cport
    global host
    runs.destroy()
    runs1.destroy()
    runs2.destroy()
    name = tk.Entry(root)
    namel = tk.Label(root, text="enter your name DO NO LEAVE BLANK" )
    cport = tk.Entry(root)
    cportl = tk.Label(root, text="enter you port for the client leave blank for 5555")
    host = tk.Entry(root)
    hostl = tk.Label(root, text="enter the host for the server")
    namel.pack()
    name.pack()
    cportl.pack()
    cport.pack()
    hostl.pack()
    host.pack()
    sumit = tk.Button(root, text="summit", command=csumit)
    sumit.pack()
def both():
    global world
    global port
    global seed
    global name
    global cport
    global host
    runs.destroy()
    runs1.destroy()
    runs2.destroy()
    world = tk.Entry(root)
    worldl = tk.Label(root, text="enter you world name leave blank for default" )
    port = tk.Entry(root)
    portl = tk.Label(root, text="enter you port for the server leave blank for 5555")
    seed = tk.Entry(root)
    seedl = tk.Label(root, text="enter the seed for the world")
    name = tk.Entry(root)
    namel = tk.Label(root, text="enter your name DO NO LEAVE BLANK" )
    cport = tk.Entry(root)
    cportl = tk.Label(root, text="enter you port for the client leave blank for 5555")
    host = tk.Entry(root)
    hostl = tk.Label(root, text="enter the host for the server")
    cportl.pack()
    cport.pack()
    namel.pack()
    name.pack()
    hostl.pack()
    host.pack()
    worldl.pack()
    world.pack()
    portl.pack()
    port.pack()
    seedl.pack()
    seed.pack()
    sumit = tk.Button(root, text="summit", command=bsumit)
    sumit.pack()
runs = tk.Button(root, text="server", command=server)
runs.pack()
runs1 = tk.Button(root, text="client", command=client)
runs1.pack()
runs2 = tk.Button(root, text="both", command=both)
runs2.pack()
root.mainloop()
