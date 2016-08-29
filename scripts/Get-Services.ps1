param(
    [Parameter(Mandatory=$true)]
    [String]$Name
)

trap {
    return [PSCustomObject]@{
        error = $Error[0]
        data = ""
    } | ConvertTo-Json -Depth 2 -Compress
}

$results = @()
foreach($i in $(Get-Service -Name $Name))
{
    $results += [PSCustomObject]@{
        name = $i.Name
        displayname = $i.DisplayName
        status = $i.Status
    }
}
return [PSCustomObject]@{
    error = ""
    data = $results
} | ConvertTo-Json -Depth 3 -Compress