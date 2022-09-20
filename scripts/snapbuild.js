const childProcess = require('child_process')
const path = require('path')
const rimraf = require('rimraf')

const repoDir = path.dirname(__dirname)

exports.buildBinary = () => {
  childProcess.execFileSync('go', ['build', '-ldflags=-s -w', './cmd/snapshot'], { cwd: repoDir, stdio: 'ignore' })
  return path.join(repoDir, process.platform === 'win32' ? 'snapshot.exe' : 'snapshot')
}

exports.removeRecursiveSync = path => {
  try {
    // Strangely node doesn't have a function to remove a directory tree.
    // Using "rm -fr" will never work on Windows because the "rm" command
    // doesn't exist. Using the "rimraf" should be cross-platform and even
    // works on Windows some of the time.
    rimraf.sync(path, { disableGlob: true })
  } catch (e) {
    // Removing stuff on Windows is flaky and unreliable. Don't fail tests
    // on CI if Windows is just being a pain. Common causes of flakes include
    // random EPERM and ENOTEMPTY errors.
    //
    // The general "solution" to this is to try asking Windows to redo the
    // failing operation repeatedly until eventually giving up after a
    // timeout. But that doesn't guarantee that flakes will be fixed so we
    // just give up instead. People that want reasonable file system
    // behavior on Windows should use WSL instead.
  }
}
