const { SourceMapConsumer } = require('source-map')
const { buildBinary, removeRecursiveSync } = require('./snapbuild')
const childProcess = require('child_process')
const path = require('path')
const fs = require('fs').promises

const snapbuildPath = buildBinary()
const testDir = path.join(__dirname, '.verify-snapshot-source-map')
let tempDirCount = 0

const toSearchBundle = {
  a0: 'a.js',
  a1: 'a.js',
  a2: 'a.js',
  b0: 'b-dir/b.js',
  b1: 'b-dir/b.js',
  b2: 'b-dir/b.js',
  c0: 'b-dir/c-dir/c.js',
  c1: 'b-dir/c-dir/c.js',
  c2: 'b-dir/c-dir/c.js',
}

const testCaseES6 = {
  'a.js': `
    import {b0} from './b-dir/b'
    function a0() { a1("a0") }
    function a1() { a2("a1") }
    function a2() { b0("a2") }
    a0()
  `,
  'b-dir/b.js': `
    import {c0} from './c-dir/c'
    export function b0() { b1("b0") }
    function b1() { b2("b1") }
    function b2() { c0("b2") }
  `,
  'b-dir/c-dir/c.js': `
    export function c0() { c1("c0") }
    function c1() { c2("c1") }
    function c2() { throw new Error("c2") }
  `,
}

const testCaseCommonJS = {
  'a.js': `
    const {b0} = require('./b-dir/b')
    function a0() { a1("a0") }
    function a1() { a2("a1") }
    function a2() { b0("a2") }
    a0()
  `,
  'b-dir/b.js': `
    const {c0} = require('./c-dir/c')
    exports.b0 = function() { b1("b0") }
    function b1() { b2("b1") }
    function b2() { c0("b2") }
  `,
  'b-dir/c-dir/c.js': `
    exports.c0 = function() { c1("c0") }
    function c1() { c2("c1") }
    function c2() { throw new Error("c2") }
  `,
}

const testCaseDiscontiguous = {
  'a.js': `
    import {b0} from './b-dir/b.js'
    import {c0} from './b-dir/c-dir/c.js'
    function a0() { a1("a0") }
    function a1() { a2("a1") }
    function a2() { b0("a2") }
    a0(b0, c0)
  `,
  'b-dir/b.js': `
    exports.b0 = function() { b1("b0") }
    function b1() { b2("b1") }
    function b2() { c0("b2") }
  `,
  'b-dir/c-dir/c.js': `
    export function c0() { c1("c0") }
    function c1() { c2("c1") }
    function c2() { throw new Error("c2") }
  `,
}

const testCaseEmptyFile = {
  'entry.js': `
    import './before'
    import {fn} from './re-export'
    import './after'
    fn()
  `,
  're-export.js': `
    // This file will be empty in the generated code, which was causing
    // an off-by-one error with the source index in the source map
    export {default as fn} from './test'
  `,
  'test.js': `
    export default function() {
      console.log("test")
    }
  `,
  'before.js': `
    console.log("before")
  `,
  'after.js': `
    console.log("after")
  `,
}

const toSearchEmptyFile = {
  before: 'before.js',
  test: 'test.js',
  after: 'after.js',
}

const testCaseNonJavaScriptFile = {
  'entry.js': `
    import './before'
    import text from './file.txt'
    import './after'
    console.log(text)
  `,
  'file.txt': `
    This is some text.
  `,
  'before.js': `
    console.log("before")
  `,
  'after.js': `
    console.log("after")
  `,
}

const toSearchNonJavaScriptFile = {
  before: 'before.js',
  after: 'after.js',
}

const testCaseUnicode = {
  'entry.js': `
    import './a'
    import './b'
  `,
  'a.js': `
    console.log('🍕🍕🍕', "a")
  `,
  'b.js': `
    console.log({𐀀: "b"})
  `,
}

const toSearchUnicode = {
  a: 'a.js',
  b: 'b.js',
}

const testCasePartialMappings = {
  // The "mappings" value is "A,Q,I;A,Q,I;A,Q,I;AAMA,QAAQ,IAAI;" which contains
  // partial mappings without original locations. This used to throw things off.
  'entry.js': `console.log(1);
console.log(2);
console.log(3);
console.log("entry");
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKIC` +
    `Aic291cmNlcyI6IFsiZW50cnkuanMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnNvb` +
    `GUubG9nKDEpXG5cbmNvbnNvbGUubG9nKDIpXG5cbmNvbnNvbGUubG9nKDMpXG5cbmNvbnNv` +
    `bGUubG9nKFwiZW50cnlcIilcbiJdLAogICJtYXBwaW5ncyI6ICJBLFEsSTtBLFEsSTtBLFE` +
    `sSTtBQU1BLFFBQVEsSUFBSTsiLAogICJuYW1lcyI6IFtdCn0=
`,
}

const testCasePartialMappingsPercentEscape = {
  // The "mappings" value is "A,Q,I;A,Q,I;A,Q,I;AAMA,QAAQ,IAAI;" which contains
  // partial mappings without original locations. This used to throw things off.
  'entry.js': `console.log(1);
console.log(2);
console.log(3);
console.log("entry");
//# sourceMappingURL=data:,%7B%22version%22%3A3%2C%22sources%22%3A%5B%22entr` +
    `y.js%22%5D%2C%22sourcesContent%22%3A%5B%22console.log(1)%5Cn%5Cnconsole` +
    `.log(2)%5Cn%5Cnconsole.log(3)%5Cn%5Cnconsole.log(%5C%22entry%5C%22)%5Cn` +
    `%22%5D%2C%22mappings%22%3A%22A%2CQ%2CI%3BA%2CQ%2CI%3BA%2CQ%2CI%3BAAMA%2` +
    `CQAAQ%2CIAAI%3B%22%2C%22names%22%3A%5B%5D%7D
`,
}

const toSearchPartialMappings = {
  entry: 'entry.js',
}

const testCaseLocalVariableSwap = {
  'entry.js': `exports['./a.js'] = require('./a.js')
exports['./b.js'] = require('./b.js')
`,
  'a.js': `'use strict';
const b = require('./b.js');

let local;
if (b.test === "z-test") {
  module.exports = b.test;
} else {
  local = process.env["a-test"];
  module.exports = local;
}
`,
  'b.js': `'use strict';
module.exports = {test: "b-test"};`
}

const toSearchLocalVariableSwap = {
  'a-test': 'a.js',
  'z-test': 'a.js',
  'b-test': 'b.js',
}

const testCaseNested = {
  'entry.js': `function nested() {
    let a
    a = require('./a')
  }
  let b, c
  b = require('./b')
  c = b.foo
  module.exports = {b, c, a: nested}
`,
  'a.js': `module.exports = {foo: "a-test"};`,
  'b.js': `'use strict';
module.exports = {foo: "b-test"};
`
}

const toSearchNested = {
  'a-test': 'a.js',
  'b-test': 'b.js',
  './b': 'entry.js',
}

async function check(kind, testCase, toSearch, { entryPoint, crlf, status }) {
  let failed = 0

  try {
    const recordCheck = (success, message) => {
      if (!success) {
        failed++
        console.error(`❌ [${kind}] ${message}`)
      }
    }

    const tempDir = path.join(testDir, `${kind}-${tempDirCount++}`)
    await fs.mkdir(tempDir, { recursive: true })

    let allPaths = []
    for (const name in testCase) {
      const tempPath = path.join(tempDir, name)
      let code = testCase[name]
      await fs.mkdir(path.dirname(tempPath), { recursive: true })
      if (crlf) code = code.replace(/\n/g, '\r\n')
      await fs.writeFile(tempPath, code)
      allPaths.push(`./${path.relative(tempDir, tempPath)}`)
    }

    const configFilePath = path.join(tempDir, 'config.json')
    await fs.writeFile(configFilePath, JSON.stringify({
      basedir: tempDir,
      entryfile: path.join(tempDir, entryPoint),
      deferred: status === 'all-deferred' ? allPaths : [],
      norewrite: status === 'all-norewrite' ? allPaths : [],
      doctor: false,
      metafile: true,
      outfile: path.join(tempDir, 'out.js'),
      sourcemap: path.join(tempDir, 'out.js.map'),
    }, null, 2))

    const args = [configFilePath, '--log-level=warning']

    await new Promise((resolve, reject) => {
      const child = childProcess.spawn(snapbuildPath, args, { cwd: tempDir, stdio: ['pipe', 'pipe', 'inherit'] })
      child.stdout.on('end', resolve)
      child.on('error', reject)
    })

    let outJs
    let outJsMap

    outJs = await fs.readFile(path.join(tempDir, 'out.js'), 'utf8')
    outJsMap = await fs.readFile(path.join(tempDir, 'out.js.map'), 'utf8')

    // // Check the mapping of various key locations back to the original source
    const checkMap = (out, map) => {
      for (const id in toSearch) {
        const outIndex = out.indexOf(`"${id}"`)
        if (outIndex < 0) throw new Error(`Failed to find "${id}" in output`)
        const outLines = out.slice(0, outIndex).split('\n')
        const outLine = outLines.length
        const outColumn = outLines[outLines.length - 1].length
        const { source, line, column } = map.originalPositionFor({ line: outLine, column: outColumn })

        const inSource = toSearch[id];
        recordCheck(source === inSource, `expected source: ${inSource}, observed source: ${source}`)

        const inJs = map.sourceContentFor(source)
        let inIndex = inJs.indexOf(`"${id}"`)
        if (inIndex < 0) inIndex = inJs.indexOf(`'${id}'`)
        if (inIndex < 0) throw new Error(`Failed to find "${id}" in input`)
        const inLines = inJs.slice(0, inIndex).split('\n')
        const inLine = inLines.length
        const inColumn = inLines[inLines.length - 1].length

        const expected = JSON.stringify({ source, line: inLine, column: inColumn })
        const observed = JSON.stringify({ source, line, column })
        recordCheck(expected === observed, `expected original position: ${expected}, observed original position: ${observed}`)

        // Also check the reverse mapping
        const positions = map.allGeneratedPositionsFor({ source, line: inLine, column: inColumn })
        recordCheck(positions.length > 0, `expected generated positions: 1, observed generated positions: ${positions.length}`)
        let found = false
        for (const { line, column } of positions) {
          if (line === outLine && column === outColumn) {
            found = true
            break
          }
        }
        const expectedPosition = JSON.stringify({ line: outLine, column: outColumn })
        const observedPositions = JSON.stringify(positions)
        recordCheck(found, `expected generated position: ${expectedPosition}, observed generated positions: ${observedPositions}`)
      }
    }

    const sources = JSON.parse(outJsMap).sources
    for (let source of sources) {
      if (sources.filter(s => s === source).length > 1) {
        throw new Error(`Duplicate source ${JSON.stringify(source)} found in source map`)
      }
    }

    const outMap = await new SourceMapConsumer(outJsMap)
    checkMap(outJs, outMap)

    // Check that every generated location has an associated original position.
    const outLines = outJs.trimRight().split('\n');

    let insideGeneratedCode = false
    for (let outLine = 0; outLine < outLines.length; outLine++) {
      if (insideGeneratedCode) {
        for (let outColumn = 0; outColumn <= outLines[outLine].length; outColumn++) {
          const { line, column } = outMap.originalPositionFor({ line: outLine + 1, column: outColumn })

          recordCheck(line !== null && column !== null, `missing location for line ${outLine+1} and column ${outColumn}`)
        }
      }

      if (/^__commonJS\[\"\S+.js/.test(outLines[outLine])) {
        insideGeneratedCode = true
      }

      if (outLines[outLine].startsWith('}')) {
        insideGeneratedCode = false
      }
    }

    if (!failed) removeRecursiveSync(tempDir)

    outMap.destroy()
  }

  catch (e) {
    console.error(`❌ [${kind}] ${e && e.message || e}`)
    failed++
  }

  return failed
}

async function main() {
  const promises = []
  for (const crlf of [false, true]) {
    for (const status of ['all-healthy', 'all-deferred', 'all-norewrite']) {
      const suffix = (crlf ? '-crlf' : '') + '-' + status
      promises.push(
        check('commonjs' + suffix, testCaseCommonJS, toSearchBundle, {
          entryPoint: 'a.js',
          crlf,
        }),
        check('es6' + suffix, testCaseES6, toSearchBundle, {
          entryPoint: 'a.js',
          crlf,
        }),
        check('discontiguous' + suffix, testCaseDiscontiguous, toSearchBundle, {
          entryPoint: 'a.js',
          crlf,
        }),
        check('empty' + suffix, testCaseEmptyFile, toSearchEmptyFile, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('non-js' + suffix, testCaseNonJavaScriptFile, toSearchNonJavaScriptFile, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('unicode' + suffix, testCaseUnicode, toSearchUnicode, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('unicode-globalName' + suffix, testCaseUnicode, toSearchUnicode, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('dummy' + suffix, testCasePartialMappings, toSearchPartialMappings, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('dummy' + suffix, testCasePartialMappingsPercentEscape, toSearchPartialMappings, {
          entryPoint: 'entry.js',
          crlf,
        }),
        check('banner-footer' + suffix, testCaseES6, toSearchBundle, {
          entryPoint: 'a.js',
          crlf,
        }),
        // Test renaming local variables
        check('local-variable-swap' + suffix, testCaseLocalVariableSwap, toSearchLocalVariableSwap, {
          entryPoint: 'entry.js',
          crlf,
          status,
        }),
        // Test renaming local variables that are nested inside of functions
        check('nested' + suffix, testCaseNested, toSearchNested, {
          entryPoint: 'entry.js',
          crlf,
          status,
        }),
      )
    }
  }

  const failed = (await Promise.all(promises)).reduce((a, b) => a + b, 0)
  if (failed > 0) {
    console.error(`❌ verify snapshot source map failed`)
    process.exit(1)
  } else {
    console.log(`✅ verify snapshot source map passed`)
    removeRecursiveSync(testDir)
  }
}

main().catch(e => setTimeout(() => { throw e }))
