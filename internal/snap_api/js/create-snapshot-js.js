'use strict'

// modified from electron-link/src/generate-snapshot-script.js

const fs = require('fs')
const path = require('path')

const snapshotScriptPath = path.join(__dirname, 'blueprint.js')

const esbuildRoot = path.resolve(__dirname, '../../../')
const metaPath = path.resolve(esbuildRoot, 'internal/snap_api/output/meta.json')
const bundlePath = path.resolve(esbuildRoot, 'internal/snap_api/output/snap.js')

const requireDefinitions = (bundle, definitions) => `
  customRequire.definitions = (function (require) {
    //
    // Start Bundle generated with esbuild
    //
    ${bundle}
    //
    // End Bundle generated with esbuild
    //

    return { ${definitions.join('\n')} 
    }
  })(customRequire)
`

function createSnapshotScript(bundlePath, metaPath, baseDir, options = {}) {
  // TODO: inject this (or a slightly modified version) into the snapshotScript

  const meta = require(metaPath)
  const bundle = fs.readFileSync(bundlePath, 'utf8')
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

  //
  // require definitions
  //
  const definitionsAssignment = 'customRequire.definitions = {}'
  const definitions = []
  for (const output of Object.values(meta.outputs)) {
    for (const input of Object.values(output.inputs)) {
      const { fullPath, replacementFunction } = input.fileInfo
      const relPath = path.relative(baseDir, fullPath)
      definitions.push(`
      './${relPath}': function (
          exports,
          module,
          __filename,
          __dirname) { ${replacementFunction}(exports, module) },`)
    }
  }

  const indentedBundle = bundle.split('\n').join('\n    ')
  const requireDefs = requireDefinitions(indentedBundle, definitions)
  snapshotScript = snapshotScript.replace(definitionsAssignment, requireDefs)

  return snapshotScript
}

module.exports = { createSnapshotScript }

const updatedScript = createSnapshotScript(
  bundlePath,
  metaPath,
  '/Volumes/d/dev/cy/perf-tr1/v8-snapshot-utils/example-minimal'
)
console.log(updatedScript)
