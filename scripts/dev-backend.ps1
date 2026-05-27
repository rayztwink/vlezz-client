Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\..\apps\backend"
try {
  go run ./cmd/rayflowd
}
finally {
  Pop-Location
}

