import { mdiAccount, mdiLogout, mdiThemeLightDark } from '@mdi/js'

import { profileRouteForRole } from '@/stores/auth.js'

export const getMenuNavBar = (role, currentUserLabel = 'Pengguna') => [
  {
    isCurrentUser: true,
    label: currentUserLabel,
    menu: [
      { icon: mdiAccount, label: 'Profil', to: profileRouteForRole(role) },
      { isDivider: true },
      { icon: mdiLogout, label: 'Log Out', isLogout: true },
    ],
  },
  {
    icon: mdiThemeLightDark,
    label: 'Light/Dark',
    isDesktopNoLabel: true,
    isToggleLightDark: true,
  },
]
