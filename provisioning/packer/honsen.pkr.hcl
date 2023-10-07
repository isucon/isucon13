packer {
  required_plugins {
    ansible = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/ansible"
    }
  }
}

variable "revision" {
  type    = string
  default = "unknown"
}

locals {
  name = "isucon13-honsen-${formatdate("YYYYMMDD-hhmm", timestamp())}"
  ami_tags = {
    Project  = "honsen"
    Family   = "isucon13-honsen"
    Name     = "${local.name}"
    Revision = "${var.revision}"
    Packer   = "1"
  }
  run_tags = {
    Project = "honsen"
    Name    = "packer-${local.name}"
    Packer  = "1"
    Ignore  = "1"
  }
}

data "amazon-ami" "ubuntu-jammy" {
  filters = {
    name                = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
    root-device-type    = "ebs"
    virtualization-type = "hvm"
  }
  most_recent = true
  owners      = ["099720109477"]
  region      = "ap-northeast-1"
}

source "amazon-ebs" "honsen" {
  ami_name    = "${local.name}"
  ami_regions = ["ap-northeast-1"]

  tags          = local.ami_tags
  snapshot_tags = local.ami_tags

  source_ami    = "${data.amazon-ami.ubuntu-jammy.id}"
  region        = "ap-northeast-1"
  instance_type = "t3.medium"

  run_tags        = local.run_tags
  run_volume_tags = local.run_tags

  ssh_interface           = "public_ip"
  ssh_username            = "ubuntu"
  temporary_key_pair_type = "ed25519"

  launch_block_device_mappings {
    volume_size = 8
    device_name = "/dev/sda1"
  }
}

build {
  sources = ["source.amazon-ebs.honsen"]

  provisioner "ansible" {
    playbook_file = "../../provisioning/ansible/application.yml"
    host_alias = "application"
    ansible_env_vars = [
      "ANSIBLE_SSH_TRANSFER_METHOD=piped",
    ]
  }
  provisioner "shell" {
    env = {
      DEBIAN_FRONTEND = "noninteractive"
    }
    inline = [
      // "cd /dev/shm",
      // "tar xf mitamae.tar.gz",
      // "cd mitamae",
      // "sudo ./setup.sh",
      // "sudo ./mitamae local roles/default.rb",

      // # install initial data and codes
      // "sudo rsync -a /dev/shm/webapp/ /home/isucon/webapp/",
      // "sudo rsync -a /dev/shm/public/ /home/isucon/public/",
      // "sudo rsync -a /dev/shm/bench/ /home/isucon/bench/",
      // "sudo rsync -a /dev/shm/data/ /home/isucon/data/",
      // "sudo tar xvf /dev/shm/initial_data.tar.gz -C /home/isucon",
      // "sudo chown -R isucon:isucon /home/isucon",

      // # reset mysql password
      // "sudo mysql -u root -p -e \"ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'root';\"",
      // "sudo cat /home/isucon/webapp/sql/admin/*.sql | mysql -uroot -proot",

      // # prepare webapp
      // "sudo ./mitamae local roles/webapp.rb",
      // "sudo -u isucon /home/isucon/webapp/sql/init.sh",

      # Remove authorized_keys for packer
      "sudo truncate -s 0 /home/ubuntu/.ssh/authorized_keys",
      "sudo truncate -s 0 /etc/machine-id",
    ]
  }
}

