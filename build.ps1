if ($PSVersionTable.PSVersion.Major -lt 6) {
  # if it's not PowerShell 6+, it's Windows (most likely)
  $output = "dumbdl.exe"
} else {
  # automatic variable $IsWindows only exists in PowerShell 6+
  if ($IsWindows) {
    $output = "dumbdl.exe"
  } else {
    $output = "dumbdl"
  }
}
go build -o $output
