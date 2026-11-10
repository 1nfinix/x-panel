import { readFileSync } from 'node:fs'
import parser from './theme-token-parser.js'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const less = read('src/styles/global.less')
const { block, tokens } = parser
const canonical = {
  light: tokens(block(less, ':root, [data-theme="light"] {')),
  dark: tokens(block(less, 'html.dark, [data-theme="dark"] {')),
}
const errors = []
const compare = (location, actual, expected) => {
  if (actual?.toLowerCase() !== expected?.toLowerCase())
    errors.push(`${location}: ${actual ?? '(missing)'} != ${expected ?? '(missing token)'}`)
}

const theme = read('src/hooks/useTheme.ts')
for (const mode of ['light', 'dark']) {
  const values = block(theme, `${mode}: {`)
  for (const [, property, value] of values.matchAll(/(\w+):\s*'([^']+)'/g)) {
    const token = property === 'modal' ? 'bg1' : property.replace(/[A-Z]/g, letter => `-${letter.toLowerCase()}`)
    compare(`useTheme.ts ${mode}.${property}`, value, canonical[mode][`--x-${token}`])
  }
}

const html = read('index.html')
for (const mode of ['light', 'dark']) {
  const meta = new RegExp(`<meta name="theme-color" content="([^"]+)" media="\\(prefers-color-scheme: ${mode}\\)"`)
  compare(`index.html ${mode} theme-color`, html.match(meta)?.[1], canonical[mode]['--x-bg0'])
  const background = new RegExp(`${mode === 'dark' ? '@media \\(prefers-color-scheme: dark\\) \\{ body \\{ background:' : 'body \\{ margin: 0; background:'}\\s*([^;]+);`)
  compare(`index.html ${mode} loading background`, html.match(background)?.[1], canonical[mode]['--x-bg0'])
}

const vite = read('vite.config.ts')
for (const key of ['theme_color', 'background_color'])
  compare(`vite.config.ts ${key}`, vite.match(new RegExp(`${key}:\\s*'([^']+)'`))?.[1], canonical.light['--x-bg0'])

if (errors.length) {
  console.error(errors.join('\n'))
  process.exitCode = 1
}
else {
  console.log('Theme tokens match global.less')
}
