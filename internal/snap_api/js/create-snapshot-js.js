'use strict'

// modified from electron-link/src/generate-snapshot-script.js

const fs = require('fs')
const path = require('path')

const snapshotScriptPath = path.join(__dirname, 'blueprint.js')

function createSnapshotScript(bundlePath, options = {}) {
  // TODO: inject this (or a slightly modified version) into the snapshotScript

  const bundle = fs.readFileSync(bundlePath)
  let snapshotScript = fs.readFileSync(snapshotScriptPath, 'utf8')

  // TODO: verify we don't need to replace `require(main)`

  //
  // Platform specifics
  //
  snapshotScript = snapshotScript.replace('processPlatform', process.platform)
  snapshotScript = snapshotScript.replace(
    'const pathSeparator = null',
    `const pathSeparator = ${JSON.stringify(path.sep)}`
  )

  //
  // Auxiliary Data
  //
  const auxiliaryData = JSON.stringify(options.auxiliaryData || {})
  const auxiliaryDataAssignment = 'var snapshotAuxiliaryData = {}'
  const auxiliaryDataAssignmentStartIndex = snapshotScript.indexOf(
    auxiliaryDataAssignment
  )
  const auxiliaryDataAssignmentEndIndex =
    auxiliaryDataAssignmentStartIndex + auxiliaryDataAssignment.length
  snapshotScript =
    snapshotScript.slice(0, auxiliaryDataAssignmentStartIndex) +
    `var snapshotAuxiliaryData = ${auxiliaryData};` +
    snapshotScript.slice(auxiliaryDataAssignmentEndIndex)

  // TODO: `require.definitions` may work differently in our case since they are already rewritten
  // to functions, i.e. the original `require` calls were replaced with function calls.
  // The main question is what is the importance of the `startRow`, `endRow` and related
  // snapshotSections?

  return snapshotScript
}

module.exports = { createSnapshotScript }

const updatedScript = createSnapshotScript(
  path.join(__dirname, '..', 'output_snap.js')
)
console.log(updatedScript)
