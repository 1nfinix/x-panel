function block(source, start) {
  const from = source.indexOf(start)
  if (from < 0)
    throw new Error(`Missing token block: ${start}`)
  const open = source.indexOf('{', from)
  return source.slice(open + 1, source.indexOf('}', open))
}

function tokens(source) {
  return Object.fromEntries([...source.matchAll(/(--x-[\w-]+):\s*([^;]+);/g)].map(([, key, value]) => [key, value.trim()]))
}

module.exports = { block, tokens }
