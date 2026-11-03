/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useCallback, useEffect, useState } from 'react'

export function useFullscreen() {
  const [isFullscreen, setIsFullscreen] = useState(false)
  // Embedded players and locked-down kiosk browsers can deny the request, so
  // the entry point is only offered when the document actually allows it.
  const [supported, setSupported] = useState(false)

  useEffect(() => {
    setSupported(document.fullscreenEnabled === true)
    const handleChange = () => {
      setIsFullscreen(document.fullscreenElement != null)
    }
    document.addEventListener('fullscreenchange', handleChange)
    return () => {
      document.removeEventListener('fullscreenchange', handleChange)
    }
  }, [])

  const toggle = useCallback(async () => {
    try {
      if (document.fullscreenElement != null) {
        await document.exitFullscreen()
        return
      }
      await document.documentElement.requestFullscreen()
    } catch {
      setSupported(false)
    }
  }, [])

  return { isFullscreen, supported, toggle }
}
