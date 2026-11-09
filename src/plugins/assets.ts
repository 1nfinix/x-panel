import '@/styles/lib/tailwind.css'
import '@/styles/global.less'
import '@fontsource-variable/noto-serif-sc'

/** Tailwind's Preflight Style Override */
function naiveStyleOverride() {
  const meta = document.createElement('meta')
  meta.name = 'naive-ui-style'
  document.head.appendChild(meta)
}

function setupAssets() {
  naiveStyleOverride()
}

export default setupAssets
