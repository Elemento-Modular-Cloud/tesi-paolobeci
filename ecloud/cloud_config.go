package ecloud

// CloudinitTemplate contains the cloud-init user-data configuration
const CloudinitTemplate = `#cloud-config

# 1. Create a user, set password, and add to the 'sudo' group
users:
  - name: root
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    lock_passwd: false
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQD035mfursjtGxrz+hs1iUA/WtmLI/mz6oF2SlYRpT6ukepG87m9e5YWgu2u7te2EtBkD1ofDURq1f97k/qoBJCHomneSL7lyUKiDtcWwjjBN89To8IPil621UCLX8whGj5me9iI3EDuP5/vRsV7tRzFVxY4lPD1Z7oz1+l0x2dEBqpAESASq3kX/ndpagql5p6lX9gxeB/qXzGY6i/ufJp7JMInkQVA1zVcsbYzTqIM28qC3VJ2mQhWp5CobNlJnruoaHmH+QPWB4DiSHgg0E4eceGZaghol1DHzkkdDMMIgnELUHems2le6vPVjzPFSoRAV5/mRDYNkL7CaiaGPmG6+EMITu/RV1ufCaxJgBYYgABEW2zsVjFjFpt7hhRd1fkd3Cx05ahMCFoCavojOWaswKBJUqoOPX0nyXKZtKpebU5jicrQ/prW7KfLeizwmmTcwMzCVQnc8W86ctNBhT80R2YnXAhPzbtmC6iJgBHkxGJObePCYP2SBwv9NSmlJKuua5ATap2XIOM9FQ1sEP/zdLe+5HuoU4gXUbrWmlbO+IbW6/kqBMOi7OpdZ1JBbsnOcbMbB2mEm99UlotvdLWTR6V/fxvIZ5JYE7ybFmsRV5tAr7YIrQ26q97hXX7HDNOqmtWQERUDImU3t+cetZ8GQGh9KMgqmCxHM/qkyfHfw== paolo.beci@gmail.com

ssh_pwauth: true

chpasswd:
  list: |
    root:password
  expire: false

# 2. Configure the network settings
hostname: myhost
network:
  version: 2
  ethernets:
    ens3:
      dhcp4: true
      nameservers:
        addresses:
          - 192.168.100.10  # IP del server DNS
        search:
          - test.k8s

# 3. Define the bash script content
write_files:
  - path: /home/root/kopsscript.sh
    permissions: "0755"
    owner: root:root
    content: |
      
      data
      
      

# # 4. Run the bash script using the 'runcmd' module
runcmd:
  - mkdir -p /mnt/disks/test.k8s--main--
  - mkdir -p /mnt/disks/test.k8s--events--
  - mkfs.ext4 /dev/vdb
  - mount /dev/vdb /mnt/disks/test.k8s--main--
  - mount /dev/vdb /mnt/disks/test.k8s--events--
  - mkdir -p /mnt/disks/test.k8s--main--/mnt
  - mkdir -p /mnt/disks/test.k8s--events--/mnt
  - sudo chown -R root:root /mnt/disks/test.k8s--main--
  - sudo chmod 755 /mnt/disks/test.k8s--main--
  - sudo chown -R root:root /mnt/disks/test.k8s--events--
  - sudo chmod 755 /mnt/disks/test.k8s--events--
  - blkid /dev/vdb1
  - [ bash, /home/root/kopsscript.sh ]`

// MetaDataTemplate contains the cloud-init meta-data configuration
const MetaDataTemplate = `instance-id: id-vm-kops`
