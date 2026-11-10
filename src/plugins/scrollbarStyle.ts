const setupScrollbarStyle = () => {
  const style = document.createElement('style')
  style.textContent = `
    html { scrollbar-width: thin; scrollbar-color: var(--x-ph) transparent; }
    ::-webkit-scrollbar { width: 8px; height: 8px; background: transparent; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { border-radius: var(--x-r-pill); background: var(--x-ph); }
    ::-webkit-scrollbar-thumb:hover { background: var(--x-icon3); }
  `
  document.head.appendChild(style)
}

export default setupScrollbarStyle
