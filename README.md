This tool will request and set temporary credentials in your shell environment variables for a given role.

```bash
$ go get -u github.com/bndw/assume-role
```

## Configuration

This tool requires a profile for each role you would like to assume in `~/.aws/config`.

Each profile **must** contain the following attributes:
* `region`
* `role_arn`
* `mfa_serial`
* `source_profile`

For example:

`~/.aws/config`:

```ini
[profile boa-igc-dev]
region = us-east-2

[profile devops@boa-igc-dev]
# Dev AWS Account.
region = us-east-2
role_arn = arn:aws:iam::1234:role/DevOps
mfa_serial = arn:aws:iam::5678:mfa/YourUser
source_profile = boa-igc-dev

[profile devops@boa-igc-prd-na2]
# Production AWS Account.
region = us-east-2
role_arn = arn:aws:iam::9012:role/DevOps
mfa_serial = arn:aws:iam::5678:mfa/YourUser
source_profile = boa-igc-dev
```

`~/.aws/credentials`:

```ini
[boa-igc-dev]
aws_access_key_id = AKIMYFAKEEXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/MYxFAKEYEXAMPLEKEY
```

Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html

## Usage

Perform an action as the given IAM role:

```bash
$ assume-role devops@boa-igc-dev aws iam get-user
```

The `assume-role` tool sets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, and `AWS_DEFAULT_REGION` environment variables and then executes the command provided.

If the role requires MFA, you will be asked for the token first:

```bash
$ assume-role devops@boa-prd-na-2 aws iam get-user
MFA code: 123456
```

If no command is provided, `assume-role` will output the temporary security credentials:

```bash
$ assume-role prod
export AWS_ACCESS_KEY_ID="ASIAI....UOCA"
export AWS_SECRET_ACCESS_KEY="DuH...G1d"
export AWS_SESSION_TOKEN="AQ...1BQ=="
export AWS_SECURITY_TOKEN="AQ...1BQ=="
export AWS_DEFAULT_REGION="us-east-2"
export AWS_REGION="us-east-2"
export ASSUMED_ROLE="devops@boa-igc-prd-na2"
# Run this to configure your shell:
# eval $(assume-role prod)
```

Or windows PowerShell:
```cmd
$env:AWS_ACCESS_KEY_ID="ASIAI....UOCA"
$env:AWS_SECRET_ACCESS_KEY="DuH...G1d"
$env:AWS_SESSION_TOKEN="AQ...1BQ=="
$env:AWS_SECURITY_TOKEN="AQ...1BQ=="
$env:AWS_REGION="us-east-2"
$env:AWS_DEFAULT_REGION="us-east-2"
$env:ASSUMED_ROLE="devops@boa-igc-prd-na2"
# Run this to configure your shell:
# assume-role.exe prod | Invoke-Expression
```
