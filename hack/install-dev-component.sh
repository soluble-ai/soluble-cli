#!/bin/bash

set -euo pipefail

install() {
    mkdir -p ~/.config/lacework/components/iac

    cat > ~/.config/lacework/components/iac/.dev <<EOF
{
    "description": "Local development version of the IAC component",
    "name": "iac",
    "version": "0.0.0-dev",
    "artifacts": [
        {
            "os": "darwin",
            "arch": "amd64"
        }
    ]
}
EOF

    cat > ~/.config/lacework/components/iac/iac <<EOF
#!/bin/bash

wd=\$(pwd)

cd $(pwd)

exec go run main.go --working-dir \$wd "\$@"
EOF

    chmod a+rx ~/.config/lacework/components/iac/iac
    echo "Development version installed, run:
"
    echo "  lacework iac version"
    echo "
to verify."
}

uninstall() {
    rm ~/.config/lacework/components/iac/.dev || true
    echo "Development version uninstalled, run:
"
    echo "  lacework install iac"
    echo "
to restore the old version"
}

if [ "${1:-}" = "uninstall" ]; then
    uninstall
elif [ "${1:-}" = "install" ]; then
    install
else
    echo "usage: $0 uninstall|install"
    echo "Installs the IAC component to run this local copy"
    exit 1
fi
