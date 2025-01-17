import { createFileRoute } from '@tanstack/react-router'
import Editor from '@monaco-editor/react'
import { useTheme } from '@/components/theme-provider'

export const Route = createFileRoute('/templates/$template')({
  component: RouteComponent,
})

function RouteComponent() {
  const { theme } = useTheme()

  return (
    <div className="flex justify-center">
      <Editor
        height="90vh"
        language="yaml"
        defaultValue={image_str}
        theme={theme == 'dark' ? 'vs-dark' : 'light'}
      />
    </div>
  )
}

const image_str = `
variant: flatcar
version: 1.0.0
passwd:
  groups:
    - name: slurm
      gid: 104632
      system: true

  users:
    - name: slurm
      system: true
      no_create_home: true
      uid: 331952
      no_user_group: true
      primary_group: slurm
      home_dir: /user/slurm
      shell: /bin/bash
{{- with $.rootpw }}
    - name: root
      password_hash: {{ . }}
{{ end }}
{{- with $.adminSSHPubKeys }}
    - name: ccradmin
      shell: /bin/bash
      groups:
        - users
        - adm
      ssh_authorized_keys:
      {{- range $i, $key := $.adminSSHPubKeys }}
        - "{{ $key }}"
      {{- end }}
{{ end }}
{{ if $.host.HasTags "gpu" }}
    - name: nvidia-persistenced
      system: true
      no_create_home: true
      home_dir: /nonexistent
      shell: /usr/sbin/nologin
{{ end }}

storage:

  disks:
    - device: {{ if $.host.HasTags "nvme" }}/dev/nvme0n1{{ else }}/dev/sda{{ end }}
      wipe_table: true
      partitions:
        - label: ROOT
          type_guid: {{ if $.host.HasTags "raid" }}be9067b9-ea49-4f15-b4f6-f36f8c9e1818{{ else }}4f68bce3-e8cd-4db1-96e7-fbcaf984b709{{ end }}

{{- if $.host.HasTags "raid" }}
    - device: /dev/nvme1n1
      wipe_table: true
      partitions:
        - label: ROOT2
          type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
    - device: /dev/nvme2n1
      wipe_table: true
      partitions:
        - label: ROOT3
          type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
    - device: /dev/nvme3n1
      wipe_table: true
      partitions:
        - label: ROOT4
          type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818

  raid:
    - name: "root_array"
      level: "stripe"
      devices:
        - "/dev/nvme0n1"
        - "/dev/nvme1n1"
        - "/dev/nvme2n1"
        - "/dev/nvme3n1"
{{ end }}

  filesystems:
    - device: {{ if $.host.HasTags "raid" }} /dev/md/root_array {{ else }}/dev/disk/by-partlabel/ROOT{{ end }}
      format: ext4
      wipe_filesystem: true
      label: ROOT

  directories:
    - path: /var/log/sssd
      mode: 0750
    - path: /opt/software/bin
      mode: 0755
    - path: /opt/software/sssd/lib
      mode: 0755
    - path: /opt/software/nss/lib
      mode: 0755
    - path: /opt/software/syslibs
      mode: 0755
    - path: /vscratch
      mode: 0755
    - path: /user
      mode: 0755
    - path: /projects
      mode: 0755
    - path: /util/software
      mode: 0755
    - path: /scratch
      mode: 0777
    - path: /var/lib/dell/srvadmin/openmanage/.ipc
      mode: 0755
    - path: /etc/cdi
      mode: 0755

{{ $arch := "x86-64" }}
{{ $march := "x86_64" }}
{{ if $.host.HasTags "grace" }}
{{ $march = "aarch64" }}
{{ $arch = "arm64" }}
{{ end }}

  links:
    - path: /etc/profile.d/z99-ccr.sh
      target: /cvmfs/soft.ccr.buffalo.edu/config/profile/bash.sh
    - path: /opt/software/bin/bash
      target: /usr/bin/bash
    - path: /opt/software/sssd/lib/libnss_sss.so.2
      target: /usr/lib/{{ $march }}-linux-gnu/libnss_sss.so.2
    - path: /opt/software/syslibs/libpam.so.0
      target: /usr/lib/{{ $march }}-linux-gnu/libpam.so.0
    - path: /opt/software/syslibs/libcap-ng.so.0
      target: /usr/lib/{{ $march }}-linux-gnu/libcap-ng.so.0
    - path: /opt/software/syslibs/libaudit.so.1
      target: /usr/lib/{{ $march }}-linux-gnu/libaudit.so.1
    - path: /opt/software/slurm/lib64/libmunge.so.2
      target: /lib/{{ $march }}-linux-gnu/libmunge.so.2
    - path: /etc/localtime
      overwrite: true
      target: /usr/share/zoneinfo/America/New_York
    - path: /opt/dell
      target: /cvmfs/soft.ccr.buffalo.edu/versions/2023.01/compat/opt/dell
{{ if $.host.HasTags "gpu" }}
    - path: /etc/systemd/system/multi-user.target.wants/nvidia-persistenced.service
      target: /usr/lib/systemd/system/nvidia-persistenced.service
    - path: /etc/systemd/system/multi-user.target.upholds/nvidia-persistenced.service
      target: /usr/lib/systemd/system/nvidia-persistenced.service
    - path: /etc/systemd/system/multi-user.target.wants/dcgm-exporter.service
      target: /usr/lib/systemd/system/dcgm-exporter.service
    - path: /etc/systemd/system/multi-user.target.upholds/dcgm-exporter.service
      target: /usr/lib/systemd/system/dcgm-exporter.service
{{ end }}
{{ if $.host.HasTags "nvswitch" }}
    - path: /etc/systemd/system/multi-user.target.wants/nvidia-fabricmanager.service
      target: /usr/lib/systemd/system/nvidia-fabricmanager.service
    - path: /etc/systemd/system/multi-user.target.upholds/nvidia-fabricmanager.service
      target: /usr/lib/systemd/system/nvidia-fabricmanager.service
{{ end }}

  files:
    - path: /etc/systemd/network/10-static.network
      mode: 0644
      contents:
        inline: |
          [Match]
          MACAddress={{ $.nic.MAC }}

          [Link]
          MTUBytes={{ or $.nic.MTU 9000 }}

          [Network]
          DNS={{ index $.nic.DNS 0 }}
          Domains={{ Join $.nic.DomainSearch " " }}
          Address={{ $.nic.IP }}
          Gateway={{ $.nic.Gateway }}
{{if not ($.host.HasTags "grace")}}
    - path: /etc/systemd/network/90-iboip.network
      mode: 0644
      contents:
        inline: |
          [Match]
          Type=infiniband
          {{ $parts := Split $.nic.AddrString "." }}
          [Network]
          Address=172.16.{{index $parts 2}}.{{index $parts 3}}/16
{{ end }}
    - path: /etc/systemd/timesyncd.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          [Time]
          NTP=ntp1.cbls.ccr.buffalo.edu ntp2.cbls.ccr.buffalo.edu ntp3.cbls.ccr.buffalo.edu
    - path: /etc/systemd/system.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          [Manager]
          DefaultLimitMEMLOCK=infinity

    - path: /etc/ssh/ssh_host_ed25519_key
      mode: 0600
      overwrite: true
      contents:
        inline: |
          -----BEGIN OPENSSH PRIVATE KEY-----
          b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
          QyNTUxOQAAACAuLIyZt6VM21z0QiVJKZEFLwePIa6UaB935N5F0bUaEQAAAIhOL9eJTi/X
          iQAAAAtzc2gtZWQyNTUxOQAAACAuLIyZt6VM21z0QiVJKZEFLwePIa6UaB935N5F0bUaEQ
          AAAECHy3eGTGU/plePF7lwGAm7JkBp2A3T7sv5C/6fm51N5C4sjJm3pUzbXPRCJUkpkQUv
          B48hrpRoH3fk3kXRtRoRAAAAAAECAwQF
          -----END OPENSSH PRIVATE KEY-----
    - path: /etc/ssh/ssh_host_ed25519_key.pub
      mode: 0644
      overwrite: true
      contents:
        inline: |
          ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIC4sjJm3pUzbXPRCJUkpkQUvB48hrpRoH3fk3kXRtRoR
    - path: /etc/ssh/ssh_host_ed25519_key-cert.pub
      mode: 0644
      contents:
        inline: |
          ssh-ed25519-cert-v01@openssh.com AAAAIHNzaC1lZDI1NTE5LWNlcnQtdjAxQG9wZW5zc2guY29tAAAAILcbxTF4RIWWLx9ySgrdU/0kmqANc4BBGh5M9qwfbF/QAAAAIC4sjJm3pUzbXPRCJUkpkQUvB48hrpRoH3fk3kXRtRoRAAAAAAAAAAAAAAACAAAAB2NvbXB1dGUAAAAAAAAAAAAAAAD//////////wAAAAAAAAAAAAAAAAAAAhcAAAAHc3NoLXJzYQAAAAMBAAEAAAIBAKOlj2x98bU7bGqu1Gv4Vk3vdQ1Osu0w7q40yHKZfiF4m7xIu4xPr1riWTAvG5h8ObQh4zQmI6MK6JJYoivfgiqPvCjeSyfjox9r8HxXsS45e7kcn/Dik958u/Ef6BkEXUVFNAd97ktVmHW57ZE8Gix7muXRjT8fqgsR8lbc7/VV3aik8/u4VHjXoFr3ivZ/l0kJTI9L7bWUO2LV/BWjPrIQZnMnxZ+hrANNEYahG87aVg6iarS0XYaQeFEPB75SzOWxjOGP5OlYbVw8qLKAQnR7L5zCOJ2Rk9V+8OC0cab1DQJdziY9waImHKXgxi2ng3gWFguRN4Mhvy3UXyWaEQOW7lxXRwH4HVw+CNU/jP2bUJh77ZOCTMIjclK0aUj1IIJe5KvLjFmfcJkAHx3JD8h9/DgFpkyKwf5EAsBO8egF8ZLju35Jhs8q4HnrreeEmsDtqPTAQDYWJguF4uIevemJxT00Glf1xDw4oPgrEzGCEpvHGgStYDfW2oNSypCeA6XxU3leqG52oKZROgCWD/tbpx9kf8wriHlJbXvEQyRVDXVMh3ykF/WBaYwCNINMMtEY3HcsIryycOYm8SSnEP8AMG6Jy97UchnaV7/3rnl4PoiKgzn0wHfRt7posboI5K/iQG7mrdyiEQV60BeJqGfHLO64oTWAiB9/iix6lvBRAAACFAAAAAxyc2Etc2hhMi01MTIAAAIASTWFnkV4GUh525c/Kk/vQq87Dsf+kYt5H2ntwZeAzAcSP+TuSNHhk8jzqZNse/zpAD0yv42jX0DMO4ovwl0KtYTjOl+Zg9aFc8YYcZOFdBuK19U//GoIfYSWBbhoVuSFUkPI2RST7P806DRtALaWc192huVPZt5TKeDpTkAwciw/N58jveoprO1EgHTM7XmeKE/MeuNdivoTyJPwAW3PghPd3eMPR/x25Yx5tA9o12HO4Rx4nbg+FaM584NedXGPuf1YehncRoARg3D5xHd6UbKfqdAIKSEQUZhm5gpInKLQoLuKoOH5Fp9Psw6oRRG5/aaCTku4HWR5eapC5dnPDwWV9jCvtg751ZbX8cxCzpfg2N94Wqt0eIOvepZj02RW8uVDm6rWTtNKJCWRnVKlmzcOsTM6RRfeKdKJSINVou4CeUjeMzd94s/fkLeB//bYh37EtEUWGjdTwiGuR9KOijLbw9j7Gy+Q63RLmjAEfROKmMiJkdeeUKr/COuI6oraWW5DLYb1hWv105uyGvn5b8G40/9o35xmWVDJ/iKkOGMMHAgbetlNVVnwRBYQdpEq+tlleYSQ9QhRfgmweRP4x204U9AnrTtHSVglioUHXndU0A8L/XlFdbfldw8zabm1VcZ93P5dHW6Y5lIw0WcZw/+AykYGEuIe1azJDEmnIQY= /etc/ssh/ssh_host_ed25519_key.pub

    - path: /etc/ssh/ssh_known_hosts
      mode: 0644
      overwrite: true
      contents:
        inline: |
          @cert-authority cpn-*,*.cbls.ccr.buffalo.edu,*.core.ccr.buffalo.edu,10.*,vortex*,cld*,srv* ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCjpY9sffG1O2xqrtRr+FZN73UNTrLtMO6uNMhymX4heJu8SLuMT69a4lkwLxuYfDm0IeM0JiOjCuiSWKIr34Iqj7wo3ksn46Mfa/B8V7EuOXu5HJ/w4pPefLvxH+gZBF1FRTQHfe5LVZh1ue2RPBose5rl0Y0/H6oLEfJW3O/1Vd2opPP7uFR416Ba94r2f5dJCUyPS+21lDti1fwVoz6yEGZzJ8WfoawDTRGGoRvO2lYOomq0tF2GkHhRDwe+UszlsYzhj+TpWG1cPKiygEJ0ey+cwjidkZPVfvDgtHGm9Q0CXc4mPcGiJhyl4MYtp4N4FhYLkTeDIb8t1F8lmhEDlu5cV0cB+B1cPgjVP4z9m1CYe+2TgkzCI3JStGlI9SCCXuSry4xZn3CZAB8dyQ/Iffw4BaZMisH+RALATvHoBfGS47t+SYbPKuB5663nhJrA7aj0wEA2FiYLheLiHr3picU9NBpX9cQ8OKD4KxMxghKbxxoErWA31tqDUsqQngOl8VN5XqhudqCmUToAlg/7W6cfZH/MK4h5SW17xEMkVQ11TId8pBf1gWmMAjSDTDLRGNx3LCK8snDmJvEkpxD/ADBuicve1HIZ2le/9655eD6IioM59MB30be6aLG6COSv4kBu5q3cohEFetAXiahnxyzuuKE1gIgff4osepbwUQ==

    - path: /etc/ssh/sshd_config.d/ccrssh.conf
      mode: 0600
      overwrite: true
      contents:
        inline: |
          HostCertificate /etc/ssh/ssh_host_ed25519_key-cert.pub
          KerberosAuthentication no
          PubkeyAuthentication yes
          UsePAM yes
          GSSAPIAuthentication no
          PasswordAuthentication no
          ChallengeResponseAuthentication no
          PermitRootLogin no
          AllowTcpForwarding no
          X11Forwarding no
          AuthorizedKeysFile .ssh/authorized_keys .ssh/authorized_keys.d/ignition

    - path: /etc/sudoers.d/90-ccradmin
      mode: 0440
      contents:
        inline: |
          ccradmin ALL=(ALL) NOPASSWD:ALL
          %sysadmin ALL=(ALL) NOPASSWD:ALL

    - path: /etc/sudoers.d/80-ipmi_exporter
      mode: 0440
      overwrite: true
      contents:
        inline: |
          prometheus ALL=(ALL) NOPASSWD: /usr/sbin/ipmimonitoring,\
                              /usr/sbin/ipmi-sensors,\
                              /usr/sbin/ipmi-dcmi,\
                              /usr/sbin/ipmi-raw,\
                              /usr/sbin/bmc-info,\
                              /usr/sbin/ipmi-chassis,\
                              /usr/sbin/ipmi-sel

    - path: /etc/default/prometheus-ipmi-exporter
      mode: 0644
      overwrite: true
      contents:
        inline: |
          ARGS="--freeipmi.path=/usr/bin --config.file=/etc/prometheus/ipmi_exporter.yml"

    - path: /etc/prometheus/ipmi_exporter.yml
      mode: 0644
      overwrite: true
      contents:
        inline: |
          # Configuration file for ipmi_exporter
          modules:
            default:
              collectors:
              - dcmi
              - bmc
              - ipmi
              - chassis
              - sel
              collector_cmd:
                ipmi: sudo
                sel: sudo
                dcmi: sudo
                chassis: sudo
                sel: sudo
                bmc: sudo
              custom_args:
                ipmi:
                - "ipmimonitoring"
                sel:
                - "ipmi-sel"
                chassis:
                - "ipmi-chassis"
                dcmi:
                - "ipmi-dcmi"
                bmc:
                - "bmc-info"

    - path: /etc/sysctl.d/75-ccr.conf
      mode: 0644
      contents:
        inline: |
          vm.overcommit_memory=1

    - path: /etc/munge/munge.key
      overwrite: true
      mode: 0400
      user:
        name: munge
      group:
        name: munge
      contents:
        source: "data:;base64,Gni8ZLdUxtFQkh1ZZgYVazSdymMIpOy6LN7gp5brwMlKO8FWD2oiUL2AOiOCcVpHvi4A7U/DEyi9BkLHojv6nlAQWEX2sAbARaB853C5u+KNDL1BhlgJ8ewPWRXNFj4gbhJVh+pp79O7RmnxZZII5BwVfE7ynUgFO8csxF6EREe5QXqalGQ2YTIcWvzq+aZgMZnc4NosEWKVOLFeX1uDiHJjwTpZ+wKHUVA1LyXTIhRdjRlzTk+9poOen4Ieqf6OkJl5yhpkjmywZt3F4Rs596hvndAKW1lLUcCEYCtzbz9gFonF3fMbUwvhrSntxeahI8gfKzDPj+zr5W3HRF91rVmxEiymfNoj1QTPsucDX9v6MB4Ly4D8ahB7FbyutUKyV9VBEN53gNvT7HouaBt/+Cla54esvu4eu2HtEG2DScDLsBB9lUq3O6b0K60pQxA21fmKGRdsCuojhpbhuCk7cttxZ67J5MaTuq028Mml8UkqQk+jF005HoC14gJ8yUpaKo91xgY+kpoNUXqBQpN7NSw+9nj5KczRjYViIv4Nu7bkc2XCdDnl0CsBq4euLG+O7a1HUWgzf3X/eySpAB3xbMKlIkdCCD94FHI1b9Tw4eo7Aw5ZRyELPtb/8lgY8G4uJkhpshGSDM6/FAoij663dRWkfblK4JI+buR5MWpat8HgpxShJOAQFTKLEElYAz9izl9eq/dEWF4UWDi8pG8xRR2siAvjeY/VWg5Yrlb0jQDcFa/lZd18IRAxwhWHpTgkJtbmq4Z6c/04u4xg2BSAt2vtUnPEX+AwboTc050rP6DR4ctVnZJBwlDy4wFWCCknPSWparSJPBwWuvNTG1PYZH32BcBFmfVAyq6mOH4CQKfdJqVpyB3nHKD2xkpP4tGqVxyq7hEAKRaqQBWBmjebE62kkxTiJJaR9r+/X0zED5Jj0iaz9VQ9rZ52F8reNG/uqWCy0MQzmQlLsfVENEdbY1WJY6VOoRlHpDz8y14qmSShexmNMW635XUW39OqpQJsPqSivMQU6etfpVSCKhhqglvQfGlTMqQOhI4Vg3/1fD19Vv9C6d5H+sQGnfim7X2VJq2fROTgu+AvGWeCYiS8IXjfstW+K1e9klxKzo3X6iJDlaTAPDo5Z62GjzBtMcWEaAp7h9XIDK1e1cDmJz4jhQGOfpCcW6xpMqXj3QQJF0RQgdF0Qm9KJZB/m1/cUNWQjtl6WnCS3ikK1zJ8RiNydN7fFDeuW/+9J+LEQDZOyepjle6eE1L6mBR5O2i8WN/e8t6g7i2RXdVwTi2eK3j+vrBE1QVagU287/agdRVSplyQZA/vsbl8oGHw1l7JchN1UFom3NBqZSIDf+Jd9aVPbA=="

    - path: /etc/krb5.keytab
      overwrite: true
      mode: 0600
      user:
        name: root
      group:
        name: root
      contents:
        source: "data:;base64,BQIAAAByAAIAFENCTFMuQ0NSLkJVRkZBTE8uRURVAARob3N0ACFjb21wdXRlLXByb2QuY2Jscy5jY3IuYnVmZmFsby5lZHUAAAABZrACJQEAEgAgz16nQhMOnkEUAZpE7KlzHrcUOZLIpr5p+W9fAs1ioZkAAAABAAAAYgACABRDQkxTLkNDUi5CVUZGQUxPLkVEVQAEaG9zdAAhY29tcHV0ZS1wcm9kLmNibHMuY2NyLmJ1ZmZhbG8uZWR1AAAAAWawAiUBABEAEAE5GGentsn+qWDzn0k5rs0AAAABAAAAYgACABRDQkxTLkNDUi5CVUZGQUxPLkVEVQAEaG9zdAAhY29tcHV0ZS1wcm9kLmNibHMuY2NyLmJ1ZmZhbG8uZWR1AAAAAWawAiUBABcAEGUDtDeUDcL9rGlfHEQW6ksAAAAB"
    - path: /etc/ipa/ca.crt
      mode: 0644
      overwrite: true
      contents:
        inline: |
          -----BEGIN CERTIFICATE-----
          MIIDrTCCApWgAwIBAgIBATANBgkqhkiG9w0BAQsFADA/MR0wGwYDVQQKDBRDQkxT
          LkNDUi5CVUZGQUxPLkVEVTEeMBwGA1UEAwwVQ2VydGlmaWNhdGUgQXV0aG9yaXR5
          MB4XDTE1MDcxNzAyNTQyMFoXDTM1MDcxNzAyNTQyMFowPzEdMBsGA1UECgwUQ0JM
          Uy5DQ1IuQlVGRkFMTy5FRFUxHjAcBgNVBAMMFUNlcnRpZmljYXRlIEF1dGhvcml0
          eTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALkJxf8QGpMxKvwq6wVR
          btOAhX4e/6IhHjZ7k5gMjiZIKhPDXjqKlu5UwEDAvr8yNfNYCeqra8/o6Uqu8J/l
          Y7BugDS6g7+juIq9jFB1JnaCOFMna+1afEYn1VlgRUn+4gP2MCZWITkJTBe4Pw+i
          P3RqC7jY00l/5ZTsvv2QQDMbf8qxPsnqZwLaZVGPCDcKXUypsOnwRsfIItC+oOoh
          cHWfrOhAJXAo60VBmlUpB4f618GY0LGpZAuEjSPIZecJbQn8N1EM24VosfC+WYoJ
          wb97JTTk2+SPSXeJK58fgBXz1m2FVbLFZwYFTqPszwjMQMj2Z+NZCF5XpTvrA3tR
          APMCAwEAAaOBszCBsDAfBgNVHSMEGDAWgBSGcv01pC9PsrhDuyDO4B4OHvJRfzAP
          BgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBxjAdBgNVHQ4EFgQUhnL9NaQv
          T7K4Q7sgzuAeDh7yUX8wTQYIKwYBBQUHAQEEQTA/MD0GCCsGAQUFBzABhjFodHRw
          Oi8vc3J2LW0xNC0zMi5jYmxzLmNjci5idWZmYWxvLmVkdTo4MC9jYS9vY3NwMA0G
          CSqGSIb3DQEBCwUAA4IBAQBVYp96snHTP6X840GvZQwevXgiwtW7xx8jSPiFs+Hk
          LURnD96jHD4I3subVwhHpDMYcsltAD8oOeKQLogOpezQQFB88GfLvnJXZHI5WBHK
          x+OgybJz+P5s8PK19UyUBB2Sj6TFNc5hcO2majzPTUabCZL40bX0onw3+PrbobhO
          321onPi3/nHnAfqvhWEk0aAB0S/KaEOfRq5r1QdHzhA2bB4pcCiyxrdbeJIskEmR
          zsup4SUK2dUsQl54qsfcWmUgSvWgHvC/7XnJ+v+FiW7K2GUTX2omzRMEL9XobI0I
          E9oLc9a3EHL9yR8ybfvp2PWHlBQThAbx3N+KqOH70cOl
          -----END CERTIFICATE-----

    - path: /etc/sssd/sssd.conf
      mode: 0600
      overwrite: true
      contents:
        inline: |
          [domain/cbls.ccr.buffalo.edu]

          id_provider = ipa
          ipa_server = _srv_, ipa-prod1.cbls.ccr.buffalo.edu
          ipa_domain = cbls.ccr.buffalo.edu
          ipa_hostname = compute-prod.cbls.ccr.buffalo.edu
          auth_provider = ipa
          chpass_provider = ipa
          access_provider = ipa
          cache_credentials = True
          ldap_tls_cacert = /etc/ipa/ca.crt
          krb5_store_password_if_offline = True
          [sssd]
          services = nss, pam

          domains = cbls.ccr.buffalo.edu
          [nss]
          homedir_substring = /home

          [pam]

          [sudo]

          [autofs]

          [ssh]

          [pac]

    - path: /etc/krb5.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          includedir /var/lib/sss/pubconf/krb5.include.d/

          [libdefaults]
            default_realm = CBLS.CCR.BUFFALO.EDU
            dns_lookup_realm = true
            rdns = false
            dns_canonicalize_hostname = false
            dns_lookup_kdc = true
            ticket_lifetime = 24h
            forwardable = true
            udp_preference_limit = 0
            default_ccache_name = KEYRING:persistent:%{uid}

          [realms]
            CBLS.CCR.BUFFALO.EDU = {
              pkinit_anchors = FILE:/etc/ipa/ca.crt
              pkinit_pool = FILE:/etc/ipa/ca.crt

            }


          [domain_realm]
            .cbls.ccr.buffalo.edu = CBLS.CCR.BUFFALO.EDU
            .core.ccr.buffalo.edu = CBLS.CCR.BUFFALO.EDU
            .dev.ccr.buffalo.edu = CBLS.CCR.BUFFALO.EDU
            cbls.ccr.buffalo.edu = CBLS.CCR.BUFFALO.EDU

    - path: /etc/containers/containers.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          [engine]
          cgroup_manager="cgroupfs"

    - path: /etc/containers/storage.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          [storage]
          driver = "overlay"
          rootless_storage_path="/scratch/$UID/containers/storage"

          [storage.options.overlay]
          mount_program = "/usr/bin/fuse-overlayfs"

    - path: /etc/security/access.conf
      mode: 0644
      overwrite: true
      contents:
        inline: |
          +:root:ALL
          +:ccradmin:ALL
          +:sysadmin:ALL
          +:ccrtech:ALL
          -:ALL:ALL

    - path: /etc/pam.d/common-session
      mode: 0644
      overwrite: true
      contents:
        inline: |
          session [default=1]     pam_permit.so
          session requisite       pam_deny.so
          session required        pam_permit.so
          session optional        pam_umask.so
          session required        pam_unix.so
          session optional        pam_sss.so

    - path: /etc/pam.d/slurm
      mode: 0644
      overwrite: true
      contents:
        inline: |
          account  required  pam_unix.so
          account  required  pam_slurm.so debug
          auth     required  pam_localuser.so
          session  required  pam_limits.so

    - path: /etc/pam.d/sshd
      mode: 0644
      overwrite: true
      contents:
        inline: |
          @include common-auth
          account    required     pam_nologin.so
          account    sufficient   pam_access.so
          @include common-account
          session [success=ok ignore=ignore module_unknown=ignore default=bad]        pam_selinux.so close
          session    required     pam_loginuid.so
          session    optional     pam_keyinit.so force revoke
          @include common-session
          session    optional     pam_motd.so  motd=/run/motd.dynamic
          session    optional     pam_motd.so noupdate
          session    optional     pam_mail.so standard noenv # [1]
          session    required     pam_limits.so
          session    required     pam_env.so # [1]
          session    required     pam_env.so user_readenv=1 envfile=/etc/default/locale
          session [success=ok ignore=ignore module_unknown=ignore default=bad]        pam_selinux.so open
          @include common-password
          -account    required     pam_slurm_adopt.so action_no_jobs=deny log_level=debug

    - path: /etc/extensions/frosty-sysext-podman-24.11.1-1-ubuntu-noble-{{ $arch }}.raw
      mode: 0644
      contents:
        source: "{{ $.endpoints.RepoURL }}/frosty/noble/frosty-sysext-podman-24.11.1-1-ubuntu-noble-{{ $arch }}.raw"

{{ if $.host.HasTags "gpu" }}
    - path: /etc/extensions/frosty-sysext-gpu-nvidia-24.11.1-1-ubuntu-noble-{{ $arch }}.raw
      mode: 0644
      contents:
        source: "{{ $.endpoints.RepoURL }}/frosty/noble/frosty-sysext-gpu-nvidia-24.11.1-1-ubuntu-noble-{{ $arch }}.raw"

    - path: /etc/udev/rules.d/60-nvidia-uvm.rules
      mode: 0644
      contents:
        inline: |
          SUBSYSTEM=="drm", KERNEL=="renderD*", GROUP="render", MODE="0666"

    - path: /etc/default/dcgm-exporter
      mode: 0644
      overwrite: true
      contents:
        inline: |
          DCGM_EXPORTER_OPTS=""
          NVIDIA_DRIVER_CAPABILITIES=all
          NVIDIA_DISABLE_REQUIRE="true"
          NVIDIA_VISIBLE_DEVICES=all
{{ end }}

{{ if $.host.HasTags "grace" }}
    - path: /etc/profile.d/z50-ccr.sh
      mode: 0644
      overwrite: true
      contents:
        inline: |
          #!/bin/bash
          export CCR_COMPAT_VERSION=2024.04
          export LMOD_SYSTEM_DEFAULT_MODULES="ccrsoft/2024.04"
{{ end }}

    - path: /etc/default/cgroup_exporter
      mode: 0644
      overwrite: true
      contents:
        inline: |
          CONFIG_PATHS=/slurm
          OPTIONS="--collect.proc --collect.proc.max-exec=100 --web.disable-exporter-metrics"

systemd:
  units:
    - name: systemd-sysext.service
      enabled: false
#      dropins:
#        - name: 10-ccr.conf
#          contents: |
#            [Unit]
#            After=systemd-udevd.service modprobe@fuse.service
    - name: ccr-sysext.service
      enabled: true
      contents: |
        [Unit]
        Description=CCR sysext
        DefaultDependencies=no
        After=local-fs.target systemd-udevd.service modprobe@fuse.service
        Before=sysinit.target
        ConditionCapability=CAP_SYS_ADMIN
        ConditionDirectoryNotEmpty=|/etc/extensions
        ConditionDirectoryNotEmpty=|/run/extensions
        ConditionDirectoryNotEmpty=|/var/lib/extensions
        ConditionDirectoryNotEmpty=|/usr/local/lib/extensions
        ConditionDirectoryNotEmpty=|/usr/lib/extensions

        [Service]
        Type=oneshot
        RemainAfterExit=yes
        ExecStart=systemd-sysext merge
        ExecStop=systemd-sysext unmerge
        ExecStartPost=systemd-tmpfiles --create /usr/lib/tmpfiles.d/podman.conf
{{ if $.host.HasTags "gpu" }}
        ExecStartPost=systemd-tmpfiles --create /usr/lib/tmpfiles.d/dcgm-exporter.conf
{{ end }}

        [Install]
        WantedBy=sysinit.target
    - name: sssd.service
      enabled: true
      dropins:
        - name: 10-ccr.conf
          contents: |
            [Unit]
            After=network-online.target
    - name: cgroup_exporter.service
      enabled: true
    - name: munge.service
      enabled: true
    - name: slurmd.service
      enabled: true
      dropins:
{{ if $.host.HasTags "gpu" }}
        - name: 20-ccr.conf
          contents: |
            [Service]
            ExecStartPre=/usr/bin/nvidia-smi
            ExecStartPost=/usr/bin/nvidia-ctk cdi generate --output=/etc/cdi/nvidia.yaml
            ExecStartPost=/usr/sbin/ldconfig
{{ end }}
        - name: 10-limits.conf
          contents: |
            [Service]
            LimitCORE=0
    - name: vscratch.mount
      enabled: true
      contents: |
        [Unit]
        Description=CCR mount vscratch
        After=network.target
        Before=remote-fs.target

        [Mount]
        What=vast.cbls.ccr.buffalo.edu:/vscratch
        Where=/vscratch
        Type=nfs
        Options=_netdev,proto=tcp,auto,async,resvport,rw,rsize=1048576,wsize=1048576,nfsvers=3,bg,nosuid,nodev,nconnect=5

        [Install]
        WantedBy=remote-fs.target
    - name: user.mount
      enabled: true
      contents: |
        [Unit]
        Description=CCR mount user
        After=network.target
        Before=remote-fs.target

        [Mount]
        What=vast.cbls.ccr.buffalo.edu:/user
        Where=/user
        Type=nfs
        Options=_netdev,proto=tcp,auto,async,resvport,rw,rsize=1048576,wsize=1048576,nfsvers=3,bg,nosuid,nodev,nconnect=5

        [Install]
        WantedBy=remote-fs.target
    - name: projects.mount
      enabled: true
      contents: |
        [Unit]
        Description=CCR mount projects
        After=network.target
        Before=remote-fs.target

        [Mount]
        What=vast.cbls.ccr.buffalo.edu:/projects
        Where=/projects
        Type=nfs
        Options=_netdev,proto=tcp,auto,async,resvport,rw,rsize=1048576,wsize=1048576,nfsvers=3,bg,nosuid,nodev,nconnect=5

        [Install]
        WantedBy=remote-fs.target
    - name: util-software.mount
      enabled: true
      contents: |
        [Unit]
        Description=CCR mount util
        After=network.target
        Before=remote-fs.target

        [Mount]
        What=vast.cbls.ccr.buffalo.edu:/util/software
        Where=/util/software
        Type=nfs
        Options=_netdev,proto=tcp,auto,async,resvport,ro,rsize=1048576,wsize=1048576,nfsvers=3,bg,nosuid,nodev,nconnect=5

        [Install]
        WantedBy=remote-fs.target
`
