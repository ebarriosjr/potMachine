# For MacOS #

## Install VirtualBox ##

https://download.virtualbox.org/virtualbox/6.0.14/VirtualBox-6.0.14-133895-OSX.dmg

## Install Vagrant ##

https://releases.hashicorp.com/vagrant/2.2.6/vagrant_2.2.6_x86_64.dmg

## Compile the potMachine binary ##

```bash
git clone https://github.com/ebarriosjr/potMachine.git
cd potMachine
GOOS=darwin go build -o pot .
```

## Move binary ##

move pot to PATH

```bash
mv pot /usr/local/bin/pot
```

## Initialize potMachine only ##

```bash
pot machine init virtualbox -ip 192.168.44.100
```

-ip could be anything you want. By default it is 192.168.44.100

-v would make the command verbose

## Initialize potMachine with minipot setup ##

```bash
pot machine init nomad -ip 192.168.44.100
```

-ip could be anything you want. By default it is 192.168.44.100

-v would make the command verbose

## Help ##

```bash
pot

Local Commands:

    machine -- Creates a local enviroment with pot jail in it
    build -- Build an image from a Potfile
    push -- Push an Pot image to a web endpoint
    login -- Log in to a Pot file server

Remote Commands:

Usage: pot command [options]

Commands:

    help	-- Show help
    version -- Show the pot version
    config  -- Show pot framework configuration
    ls/list	-- List of the installed pots
    show	-- Show pot information
    info    -- Print minimal information on a pot
    top     -- Run the unix top in the pot
    ps      -- Show running pots
    init	-- Initialize the ZFS layout
    de-init	-- Deinstall pot from your system
    vnet-start -- Start the vnet configuration
    create-base	-- Create a new base image
    create-fscomp -- Create a new fs component
    create-private-bridge -- Create a new private bridge
    create -- Create a new pot (jail)
    clone -- Clone a pot creating a new one
    clone-fscomp - Clone a fscomp
    rename -- Rename a pot
    destroy -- Destroy a pot
    prune   -- Destroy not running prunable pots
    copy-in -- Copy a file or a directory into a pot
    mount-in -- Mount a directory, a zfs dataset or a fscomp into a pot
    add-dep -- Add a dependency
    set-rss -- Set a resource constraint
    get-rss -- Get the current resource usage
    set-cmd -- Set the command to start the pot
    set-env -- Set environment variabls inside a pot
    set-hosts -- Set etc/hosts entries inside a pot
    set-attr -- Set a pot's attribute
    get-attr -- Get a pot's attribute
    export-ports -- Let export tcp ports
    start -- Start a jail (pot)
    stop -- Stop a jail (pot)
    term -- Start a terminal in a pot
    run -- Start and open a terminal in a pot
    snap/snapshot -- Take a snapshot of a pot
    rollback/revert -- Restore the last snapshot
    purge-snapshots -- Remove old/all snapshots
    export -- Export a pot to a file
    import -- Import a pot from a file or a URL
    prepare -- Import and prepare a pot - designed for jail orchestrator
    update-config -- Update the configuration of a pot
```

## Potfile example for nginx ##

```bash
FROM 12.0
NAME tcs-nginx
COPY index.html /usr/local/www/nginx-dist/index.html
RUN sed -i '' 's/quarterly/latest/' /etc/pkg/FreeBSD.conf
RUN pkg install -y nginx
RUN pkg clean -a -y
FLAVOUR slim
CMD nginx
```

## Commands available on Potfile ##

FROM

NAME

COPY

ADD

ENV

RUN

FLAVOUR

EXPOSE

MEMORY

CPU

ARG

CMD
