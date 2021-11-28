package apt

const essentialPackagesToBeRemovedOutput = `Reading package lists...
Building dependency tree...
Reading state information...
The following packages were automatically installed and are no longer required:
  distro-info laptop-detect
Use 'apt autoremove' to remove them.
The following packages will be REMOVED
  apt apt-utils tasksel tasksel-data ubuntu-advantage-tools ubuntu-minimal
  ubuntu-server update-notifier-common
The following NEW packages will be installed
  apt-transport-https
WARNING: The following essential packages will be removed.
This should NOT be done unless you know exactly what you are doing!
  apt
0 to upgrade, 1 to newly install, 8 to remove and 0 not to upgrade.
E: Essential packages were removed and -y was used without --allow-remove-essential.`

const unmetDependenciesOutput = `Reading package lists...
Building dependency tree...
Reading state information...
Some packages could not be installed. This may mean that you have
requested an impossible situation or if you are using the unstable
distribution that some required packages have not yet been created
or been moved out of Incoming.
The following information may help to resolve the situation:

The following packages have unmet dependencies.
 kubeadm : Depends: kubelet (>= 1.13.0) but it is not installable
           Depends: kubectl (>= 1.13.0) but it is not installable
E: Unable to correct problems, you have held broken packages.`

const newerVersionInstalledOutput = `Reading package lists...
Building dependency tree...
Reading state information...
The following packages will be DOWNGRADED:
  libseccomp2
0 to upgrade, 0 to newly install, 1 to downgrade, 0 to remove and 0 not to upgrade.
E: Packages were downgraded and -y was used without --allow-downgrades.`

const unableToLocatePackage = `Reading package lists... Done
Building dependency tree       
Reading state information... Done
E: Unable to locate package sfsdfsdfsdf`
