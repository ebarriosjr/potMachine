# Description #

Pot Machine creates a local vagrant VM with FreeBSD12 running on zfs.
This allows developers to test local pots (https://github.com/pizzamig/pot) on linux or MacOS.
This implementation tries to mimic as much as possible the model implemented by docker.
This means you can have Potfiles where you write the configuration for your pot,
and running a pot build will create a freeBSD jail based on this "config".

There is also the possibility to create a miniPot enviroment that includes Pot, nomad and consul by default.

This program will create a folder on ~/.pot that will contain all the configuration and exports made wit it. This ~/.pot folder will be mounted inside the vagrant vm under /vagrant

# For MacOS #

### Install VirtualBox ###

<https://download.virtualbox.org/virtualbox/6.0.14/VirtualBox-6.0.14-133895-OSX.dmg>

### Install Vagrant ###

<https://releases.hashicorp.com/vagrant/2.2.6/vagrant_2.2.6_x86_64.dmg>

### Compile the potMachine binary ###

```bash
git clone https://github.com/ebarriosjr/potMachine.git
cd potMachine
GOOS=darwin go build -o pot .
```

# For Linux #

### Install VirtualBox ###

<https://www.virtualbox.org/wiki/Linux_Downloads>

### Install Vagrant ###

https://releases.hashicorp.com/vagrant/2.2.6/vagrant_2.2.6_linux_amd64.zip

### Compile the potMachine binary ###

```bash
git clone https://github.com/ebarriosjr/potMachine.git
cd potMachine
GOOS=linux go build -o pot .
```

# Move binary #

For both platforms (Linux/MacOS) the binarie need to be moved to a folder inside the user $PATH.

Example:

```bash
mv pot /usr/local/bin/pot
```

# Initialize potMachine #

This command will start a Vagrant VM running pot without nomad, consul or traeffic.

```bash
pot machine init virtualbox
```

### Command Options ###

```bash
Options:
  -ip -- Assigns an IP to the potMachine
  -v -- Verbose
```

# Potfile #

Potfile is an configuration file si,ilar to Dockerfile. With this file a build of a pot can be triggered.

In order to create a Pot with a Potfile the command `pot build .` need to be executed. You can also tag your build to export the pot in one command. 

For example:

```bash
pot build -t fileserver.example/pot/nginx:0.1
```

This command will create the pot and export the zfs dataset in a compres xz format to ~/.pot/exports

In order to push the pot dataset you need to run the following command:

```bash
pot push -t artifactory.local/artifactory/generic-local/pot/nginx:0.1
```

If your file server requires authentication you can give Pot this information by running:

```bash
pot login -u username --password-stdin fileserver.example
```

This information will be save to `~/.pot/config.json` and will be used on every push in that domain.


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

Command | Description | example
--- | --- | ---
`FROM` | Base FreeBSD OS to run inside the jail| FROM 12.1
`NAME` | Name for the pot jail. | NAME nginx
`COPY` | Copy local files to the jail after running all the command on the RUN stanza | COPY index.html /usr/local/www/nginx-dist/index.html
`ADD` | Downloads remote file to the pot jail | ADD <https://fileserver.com/test.rar>
`ARG` | ENV variable that get added inside the `creation` process of the pot jail | ARG VAR=VALUE
`RUN` | Command to be executed on `creation` of the pot jail | RUN pkg install -y nginx
`FLAVOUR` | Predifined or user created scripts to apply to a pot after RUN is done | FLAVOUR slim
`EXPOSE` | Tells the Pot which port should be exposed | EXPOSE 80
`MEMORY` | Memory limitation for the pot jail | MEMORY 1024M
`CPU` | Number of cores assing to this pot jail | CPU 2
`ENV` | Adds enviroment variables inside the running pot jail | ENV VAR=VALUE
`CMD` | Array of commands that will be executed on pot start |CMD ["nginx","-g","'daemon off;'"]

## Help ##

```bash
pot -h

Local Commands:

    machine -- Creates a local enviroment with pot jail in it
    build -- Build an image from a Potfile
    push -- Push an Pot image to a web endpoint
    login -- Log into a Pot file server

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

# Bonus - Mesh #

If you want to run a service mesh on your local machine you can use Minipot.

Minipot is a program that runs nomad, consul, traeffic and pot in a single Vagrant VM. This will give you and enviroment to test and develop your applications in a local enviroment.

```bash
pot machine init nomad
```

### Command Options ###

```bash
Options:
  -ip -- Assigns an IP to the potMachine
  -v -- Verbose
```

# EXPERIMENTAL MacOS #

## Xhyve Support for MacOS ##

### Install xhyve ###

```bash
brew install xhyve
```

### Compile the potMachine binary ###

```bash
git clone https://github.com/ebarriosjr/potMachine.git
cd potMachine
GOOS=darwin go build -o pot .
```