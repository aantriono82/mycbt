export const gradientBgBase = 'bg-linear-to-tr'
export const gradientBgPurplePink = `${gradientBgBase} from-purple-400 via-pink-500 to-red-500`
export const gradientBgDark = `${gradientBgBase} from-slate-700 via-slate-900 to-slate-800`
export const gradientBgPinkRed = `${gradientBgBase} from-pink-400 via-red-500 to-yellow-500`

export const colorsBgLight = {
  white: 'bg-white text-slate-900 border-slate-200',
  light: 'bg-slate-50 text-slate-900 border-slate-200 dark:bg-slate-900 dark:text-slate-100 dark:border-slate-800',
  contrast: 'bg-slate-900 text-white dark:bg-white dark:text-slate-900',
  success: 'bg-emerald-600 border-emerald-600 text-white dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20',
  danger: 'bg-rose-600 border-rose-600 text-white dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20',
  warning: 'bg-amber-500 border-amber-500 text-white dark:bg-amber-500/10 dark:text-amber-400 dark:border-amber-500/20',
  info: 'bg-indigo-600 border-indigo-600 text-white dark:bg-indigo-500/10 dark:text-indigo-400 dark:border-indigo-500/20',
  purple: 'bg-violet-600 border-violet-600 text-white dark:bg-violet-500/10 dark:text-violet-400 dark:border-violet-500/20',
}

export const colorsText = {
  white: 'text-slate-900 dark:text-white',
  light: 'text-slate-600 dark:text-slate-300 font-medium',
  contrast: 'text-white dark:text-slate-900',
  success: 'text-emerald-700 dark:text-emerald-400 font-bold',
  danger: 'text-rose-700 dark:text-rose-400 font-bold',
  warning: 'text-amber-700 dark:text-amber-400 font-bold',
  info: 'text-indigo-700 dark:text-indigo-400 font-bold',
  purple: 'text-violet-700 dark:text-violet-400 font-bold',
}

export const colorsOutline = {
  white: [colorsText.white, 'border-gray-100'],
  light: [colorsText.light, 'border-gray-100'],
  contrast: [colorsText.contrast, 'border-gray-900 dark:border-slate-100'],
  success: [colorsText.success, 'border-emerald-500'],
  danger: [colorsText.danger, 'border-rose-500'],
  warning: [colorsText.warning, 'border-amber-500'],
  info: [colorsText.info, 'border-indigo-500'],
  purple: [colorsText.purple, 'border-violet-500'],
}

export const getButtonColor = (color, isOutlined, hasHover, isActive = false) => {
  const colors = {
    ring: {
      white: 'ring-gray-200 dark:ring-gray-500',
      whiteDark: 'ring-gray-200 dark:ring-gray-500',
      lightDark: 'ring-gray-200 dark:ring-gray-500',
      contrast: 'ring-gray-300 dark:ring-gray-400',
      success: 'ring-emerald-300 dark:ring-emerald-700',
      danger: 'ring-rose-300 dark:ring-rose-700',
      warning: 'ring-amber-300 dark:ring-amber-700',
      info: 'ring-indigo-300 dark:ring-indigo-700',
      purple: 'ring-violet-300 dark:ring-violet-700',
    },
    active: {
      white: 'bg-gray-100',
      whiteDark: 'bg-gray-100 dark:bg-slate-800',
      lightDark: 'bg-gray-200 dark:bg-slate-700',
      contrast: 'bg-gray-700 dark:bg-slate-100',
      success: 'bg-emerald-700 dark:bg-emerald-600',
      danger: 'bg-rose-700 dark:bg-rose-600',
      warning: 'bg-amber-700 dark:bg-amber-600',
      info: 'bg-indigo-700 dark:bg-indigo-600',
      purple: 'bg-violet-700 dark:bg-violet-600',
    },
    bg: {
      white: 'bg-white text-black',
      whiteDark: 'bg-white text-black dark:bg-slate-900 dark:text-white',
      lightDark: 'bg-gray-100 text-black dark:bg-slate-800 dark:text-white',
      contrast: 'bg-gray-800 text-white dark:bg-white dark:text-black',
      success: 'bg-emerald-600 dark:bg-emerald-500 text-white',
      danger: 'bg-rose-600 dark:bg-rose-500 text-white',
      warning: 'bg-amber-600 dark:bg-amber-500 text-white',
      info: 'bg-indigo-600 dark:bg-indigo-500 text-white',
      purple: 'bg-violet-600 dark:bg-violet-500 text-white',
    },
    bgHover: {
      white: 'hover:bg-gray-100',
      whiteDark: 'hover:bg-gray-100 dark:hover:bg-slate-800',
      lightDark: 'hover:bg-gray-200 dark:hover:bg-slate-700',
      contrast: 'hover:bg-gray-700 dark:hover:bg-slate-100',
      success:
        'hover:bg-emerald-700 hover:border-emerald-700 dark:hover:bg-emerald-600 dark:hover:border-emerald-600',
      danger:
        'hover:bg-rose-700 hover:border-rose-700 dark:hover:bg-rose-600 dark:hover:border-rose-600',
      warning:
        'hover:bg-amber-700 hover:border-amber-700 dark:hover:bg-amber-600 dark:hover:border-amber-600',
      info: 'hover:bg-indigo-700 hover:border-indigo-700 dark:hover:bg-indigo-600 dark:hover:border-indigo-600',
      purple: 'hover:bg-violet-700 hover:border-violet-700 dark:hover:bg-violet-600 dark:hover:border-violet-600',
    },
    borders: {
      white: 'border-white',
      whiteDark: 'border-white dark:border-slate-900',
      lightDark: 'border-gray-100 dark:border-slate-800',
      contrast: 'border-gray-800 dark:border-white',
      success: 'border-emerald-600 dark:border-emerald-500',
      danger: 'border-rose-600 dark:border-rose-500',
      warning: 'border-amber-600 dark:border-amber-500',
      info: 'border-indigo-600 dark:border-indigo-500',
      purple: 'border-violet-600 dark:border-violet-500',
    },
    text: {
      contrast: 'dark:text-slate-100',
      success: 'text-emerald-600 dark:text-emerald-500',
      danger: 'text-rose-600 dark:text-rose-500',
      warning: 'text-amber-600 dark:text-amber-500',
      info: 'text-indigo-600 dark:text-indigo-500',
      purple: 'text-violet-600 dark:text-violet-500',
    },
    outlineHover: {
      contrast:
        'hover:bg-gray-800 hover:text-gray-100 dark:hover:bg-slate-100 dark:hover:text-black',
      success:
        'hover:bg-emerald-600 hover:text-white dark:hover:text-white dark:hover:border-emerald-600',
      danger:
        'hover:bg-rose-600 hover:text-white dark:hover:text-white dark:hover:border-rose-600',
      warning:
        'hover:bg-amber-600 hover:text-white dark:hover:text-white dark:hover:border-amber-600',
      info: 'hover:bg-indigo-600 hover:text-white dark:hover:text-white dark:hover:border-indigo-600',
      purple: 'hover:bg-violet-600 hover:text-white dark:hover:text-white dark:hover:border-violet-600',
    },
  }

  if (!colors.bg[color]) {
    return color
  }

  const isOutlinedProcessed = isOutlined && ['white', 'whiteDark', 'lightDark'].indexOf(color) < 0

  const base = [colors.borders[color], colors.ring[color]]

  if (isActive) {
    base.push(colors.active[color])
  } else {
    base.push(isOutlinedProcessed ? colors.text[color] : colors.bg[color])
  }

  if (hasHover) {
    base.push(isOutlinedProcessed ? colors.outlineHover[color] : colors.bgHover[color])
  }

  return base
}

