import { createContext, useCallback, useContext, useMemo, useState } from 'react'
import { AlertCircle, CheckCircle2, Info, TriangleAlert, X } from 'lucide-react'

const ToastContext = createContext(null)

const iconMap = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: TriangleAlert,
  info: Info
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const removeToast = useCallback((id) => {
    setToasts((items) => items.filter((item) => item.id !== id))
  }, [])

  const pushToast = useCallback(
    (type, message) => {
      const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
      setToasts((items) => [...items, { id, type, message }])
      window.setTimeout(() => removeToast(id), 2800)
    },
    [removeToast]
  )

  const value = useMemo(
    () => ({
      success: (message) => pushToast('success', message),
      error: (message) => pushToast('error', message),
      warning: (message) => pushToast('warning', message),
      info: (message) => pushToast('info', message)
    }),
    [pushToast]
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toast-viewport" aria-live="polite" aria-atomic="true">
        {toasts.map((toast) => {
          const Icon = iconMap[toast.type] || Info

          return (
            <div className={`toast toast-${toast.type}`} key={toast.id}>
              <Icon size={18} />
              <span>{toast.message}</span>
              <button
                className="toast-close"
                type="button"
                aria-label="关闭提示"
                onClick={() => removeToast(toast.id)}
              >
                <X size={16} />
              </button>
            </div>
          )
        })}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const value = useContext(ToastContext)
  if (!value) {
    throw new Error('useToast must be used inside ToastProvider')
  }

  return value
}
