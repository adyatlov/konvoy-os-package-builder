# konvoy-os-package-builder
CMD tool that composes an OS packages bundle for air-gapped Konvoy installations

## How to use
1. Build the tool:

    ```
    ./build_linux
    ```
    
2. Find a machine connected to the Internet. Install the OS on that machine from the same image you used to provision the cluster nodes.
3. Copy the binary `konvoy-os-package-builder` and the original OS package bundle `konvoy_v1.8.3_amd64_debs.tar.gz` to the same directory on that machine.
4. Rename `konvoy_v1.8.3_amd64_debs.tar.gz` to `backup_konvoy_v1.8.3_amd64_debs.tar.gz`.
5. Launch the tool: `./konvoy-os-package-builder`. The output should look like [this](notes/res.txt).
6. If the command runs successfully it creates the new `konvoy_v1.8.3_amd64_debs.tar.gz` file.
7. In the directory of the Konvoy distributive replace the old OS package bundle `konvoy_v1.8.3_amd64_debs.tar.gz` with the new one.

## Limitations
At this moment, the tool supports APT (`.deb`) packages only.
