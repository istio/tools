function downloadCSV() {
    const shell = require('shelljs')
    shell.exec('chmod u+x ../script/download_gcs.sh')
    shell.exec('../script/download_gcs.sh')
}

