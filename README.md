# Packer Plugin for Anka

This is a [Packer Builder] for building images that work with [Veertu Anka], a macOS virtualization tool.

Note that this builder does not manage images. Once it creates an image, it is up to you to use it or delete it.

### v2.0.0 Breaking Changes

* Plugin will only work with Packer v1.7 or later.
* Plugin has been renamed from `packer-builder-veertu-anka` to `packer-plugin-veertu-anka`.
* Builder has been renamed from `veertu-anka` to `veertu-anka-vm-clone` and `veertu-anka-vm-create`.

### Compatibility

Packer Version | Veertu Anka Plugin Version
--- | ---
1.7.x and above | 2.0.0

## Installing from Binary

1. [Install Packer v1.7 or newer](https://www.packer.io/downloads)
2. [Install Veertu Anka v2.3.1 or newer](https://veertu.com/download-anka-build/)
3. Download the [latest release](https://github.com/veertuinc/packer-builder-veertu-anka/releases) for your host environment
4. Unzip the plugin binaries to a location where Packer will detect them at run-time, such as any of the following:
    * The directory where the packer binary is.
    * The `~/.packer.d/plugins` directory.
    * The current working directory.
5. Change to a directory where you have packer templates, and run as usual.

## Usage

The most basic json file you can build from is:

```json
{
  "builders": [
    {
      "installer_app": "/Applications/Install macOS Big Sur.app",
      "type": "veertu-anka-vm-create",
      "vm_name": "macos-big-sur"
    }
  ],
  "post-processors": [
    {
      "type": "veertu-anka-registry-push",
      "tag": "latest"
    }
  ]
}
```

This will create a base VM template using the `.app` you specified in `installer_app` with a name like `anka-packer-base-{macOSVersion}`. Once the base VM template is created, it will create a clone from it (that shares the underlying layers from the base VM template, minimizing the amount of disk space used). Once the VM has been successfully created, it will push that VM to your default registry with the `latest` tag.

> When using `installer_app`, you can modify the base VM default resource values with `disk_size`, `ram_size`, and `cpu_count`. Otherwise, defaults will be used (see "Configuration" section).

You can also skip the creation of the base VM template and use an existing VM template (`10.15.6`):

```json
{
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "sleep 5",
        "echo hello world",
        "echo llamas rock"
      ]
    }
  ],
  "builders": [{
    "type": "veertu-anka-vm-clone",
    "source_vm_name": "10.15.6",
    "vm_name": "macos-from-source_10.15.6"
  }]
}
```

Or, create a variable inside for the `source_vm_name` and then run: `packer build -var 'source_vm_name=10.15.6' examples/macos-catalina-existing.json`.

> The `installer_app` is ignored if you've specified `source_vm_name` and it does not exist already

This will clone `10.15.6` to a new VM and, if there are differences from the base VM, modify CPU, RAM, and DISK.

> Check out the [examples directory](./examples) to see how port-forwarding and other options are used

## Builders 

### veertu-anka-vm-create

#### Required Configuration

* `installer_app` (String)

The path to a macOS installer. This process takes about 20 minutes. The resulting VM template name will be `{{vm_name}}-{{macOSVersion}}`. macOSVersion is pulled from the installer app.

* `type` (String)

Must be `veertu-anka-vm-create`.

* `vm_name` (String)

The name for the VM that is created. One is generated if not provided (`anka-packer-{10RandomCharacters}`).

#### Optional Configuration

* `anka_password` (String)

Sets the password for the vm. Can also be set with `ANKA_DEFAULT_PASSWD` env var. Defaults to `admin`.

* `anka_user` (String)

Sets the username for the vm. Can also be set with `ANKA_DEFAULT_USER` env var. Defaults to `anka`.

* `boot_delay` (String)

The time to wait before running packer provisioner commands, defaults to `10s`.

* `cpu_count` (String)

The number of CPU cores, defaults to `2`.

* `disk_size` (String)

The size in "[0-9]+G" format, defaults to `25G`.

> We will automatically resize the internal disk for you by executing: `diskutil apfs resizeContainer disk1 0`

* `hw_uuid` (String)

The Hardware UUID you wish to set (usually generated with `uuidgen`).

* `port_forwarding_rules` (Struct)

> If port forwarding rules are already set and you want to not have them fail the packer build, use `packer build --force`

```json
  "builders": [{
    "type": "veertu-anka-vm-clone",
    "source_vm_name": "anka-packer-base-10.15.7",
    "port_forwarding_rules": [
      {
        "port_forwarding_guest_port": 80,
        "port_forwarding_host_port": 12345,
        "port_forwarding_rule_name": "website"
      },
      {
        "port_forwarding_guest_port": 8080
      }
    ],
    "vm_name": "macos-from-packer-base_10.15.7"
  }]
```

* `ram_size` (String)

The size in "[0-9]+G" format, defaults to `2G`.

* `stop_vm` (Boolean)

Whether or not to stop the vm after it has been created, defaults to false.

### veertu-anka-vm-clone

#### Required Configuration

* `source_vm_name` (String)

The VM to clone for provisioning, either stopped or suspended.

* `type` (String)

Must be `veertu-anka-vm-clone`.

* `vm_name` (String)

The name for the VM that is created. One is generated if not provided (`anka-packer-{10RandomCharacters}`).

#### Optional Configuration

* `always_fetch` (Boolean)

Always pull the source VM from the registry. Defaults to false.

* `boot_delay` (String)

The time to wait before running packer provisioner commands, defaults to `10s`.

* `cacert` (String)

Path to a CA Root certificate.

* `cert` (String)

Path to your node certificate (if certificate authority is enabled).

* `cpu_count` (String)

The number of CPU cores, defaults to `2`.

* `disk_size` (String)

The size in "[0-9]+G" format, defaults to `25G`.

> We will automatically resize the internal disk for you by executing: `diskutil apfs resizeContainer disk1 0`

* `insecure` (Boolean)

Skip TLS verification.

* `key` (String)

Path to your node certificate key if the client/node certificate doesn't contain one.

* `hw_uuid` (String)

The Hardware UUID you wish to set (usually generated with `uuidgen`).

* `port_forwarding_rules` (Struct)

> If port forwarding rules are already set and you want to not have them fail the packer build, use `packer build --force`

```json
  "builders": [{
    "type": "veertu-anka-vm-clone",
    "source_vm_name": "anka-packer-base-10.15.7",
    "port_forwarding_rules": [
      {
        "port_forwarding_guest_port": 80,
        "port_forwarding_host_port": 12345,
        "port_forwarding_rule_name": "website"
      },
      {
        "port_forwarding_guest_port": 8080
      }
    ],
    "vm_name": "macos-from-packer-base_10.15.7"
  }]
```

* `ram_size` (String)

The size in "[0-9]+G" format, defaults to `2G`.

* `registry-path` (String)

The registry URL (will use your default configuration if not set).

* `remote` (String)

The registry name (will use your default configuration if not set).

* `source_vm_tag` (String)

Specify the tag of the VM we want to clone instead of using the default.

* `stop_vm` (Boolean)

Whether or not to stop the vm after it has been created, defaults to false.

* `update_addons` (Boolean)

Update the vm addons. Defaults to false.

## Post Processors

### veertu-anka-registry-push

#### Required Configuration

* `type` (String)

Must be `veertu-anka-registry-push`

#### Optional Configuration

* `cacert` (String)

Path to a CA Root certificate.

* `cert` (String)

Path to your node certificate (if certificate authority is enabled).

* `description` (String)

The description of the tag.

* `insecure` (Boolean)

Skip TLS verification.

* `key` (String)

Path to your node certificate key if the client/node certificate doesn't contain one.

* `local` (Boolean)

Assign a tag to your local template and avoid pushing to the Registry.

* `registry-path` (String)

The registry URL (will use your default configuration if not set).

* `remote` (String)

The registry name (will use your default configuration if not set).

* `remote-vm` (String)

The name of a registry template you want to push the local template onto.

* `tag` (String)

The name of the tag to push (will default as 'latest' if not set).

## Build Variables

Packer allows for the exposure of build variables which connects information related to the artifact that was built. Those variables can then be accessed by `post-processors` and `provisioners`.

The variables we expose are:

* `VMName`: name of the artifact vm
* `OSVersion`: OS version from which the artifact was created 
  * eg. 10.15.7
* `DarwinVersion`: Darwin version that is compatible with the current OS version
  * eg. 19.6.0

```json
{
  "variables": {
    "source_vm_name": "anka-packer-base-11.2-16.4.06"
  },
  "builders": [{
    "type": "",
    "source_vm_name": "{{ user `source_vm_name` }}",
    "vm_name": "anka-macos-from-{{ user `source_vm_name` }}"
  }],
  "provisioners": [
    {
      "type": "shell",
      "environment_vars": [
        "VMNAME={{ build `VMName`}}",
        "OSVERSION={{ build `OSVersion` }}",
        "DARWINVERSION={{ build `DarwinVersion` }}"
      ],
      "inline": [
        "echo $VMNAME was cloned with Mac $OSVERSION and compatible Darwin Version $DARWINVERSION"
      ]
    }
  ]
}
```

## Development

You will need a recent golang installed and setup. See `go.mod` for which version is expected.

We use [gomock](https://github.com/golang/mock) to quickly and reliably mock our interfaces for testing. This allows us to easily test when we expect logic to be called without having to rewrite golang standard library functions with custom mock logic. To generate one of these mocked interfaces, installed the mockgen binary by following the link provided.

```bash
mockgen -source=client/client.go -destination=mocks/client_mock.go -package=mocks
```

### Testing

To test a basic vm creation, run:

```bash
make packer-test
```

To test the post processor you will need an active vpn connection that can reach an anka registry. You can setup an anka registry by either adding the registry locally with:

```bash
anka registry add <registry_name> <registry_url>
```

-or-

You can setup the `create-from-installer-with-post-processing.json` with the correct registry values and update the make target `packer-test` to use that json file and run:

```bash
make packer-test
```

[Packer Builder]: https://www.packer.io/docs/extending/custom-builders.html
[Veertu Anka]: https://veertu.com/
