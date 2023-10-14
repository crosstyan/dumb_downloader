$IsWin = $false
if ($PSVersionTable.PSVersion.Major -lt 6) {
  # if it's not PowerShell 6+, it's Windows (most likely)
  $IsWin = $true
} else {
  # automatic variable $IsWindows only exists in PowerShell 6+
  if ($IsWindows) {
    $IsWin = $true
  } else {
    $IsWin = $false
  }
}

if ($IsWin) {
  $output = "dumbdl.exe"
} else {
  $output = "dumbdl"
}

if ($IsWin) {
  $swag = "swag.exe"
} else {
  $swag = "swag"
}

$api_src = "cmd/serve.go"
if ($null -eq (Get-Command $swag -ErrorAction SilentlyContinue)) { 
  Write-Host "failed to find $swag" -ForegroundColor Red
  Write-Host "please install swag by running 'go get -u github.com/swaggo/swag/cmd/swag'" -ForegroundColor Red
}
swag init -g $api_src
go build -o $output
