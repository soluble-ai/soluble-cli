#!/bin/bash

set -euo pipefail

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
