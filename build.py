#!/usr/bin/env python3
import os
import platform
import subprocess
from pathlib import Path
from shutil import which

script_path = os.path.dirname(os.path.realpath(__file__))

def main():
    is_windows = platform.system() == 'Windows'

    output = 'dumbdl.exe' if is_windows else 'dumbdl'
    swag = which('swag')
    if swag is None:
        print("swag not found")
        print("please install swag by running 'go get -u github.com/swaggo/swag/cmd/swag'")
        return

    api_src = 'cmd/serve.go'
    subprocess.run(args=[swag, 'init', '-g', api_src],
                   shell=True, stderr=subprocess.STDOUT, cwd=script_path)
    subprocess.run(args=['go', 'build', '-o', output],
                   shell=True, stderr=subprocess.STDOUT, cwd=script_path)

if __name__ == '__main__':
    main()
