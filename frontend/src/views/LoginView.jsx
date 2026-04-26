import { useState } from 'react'
import { Eye, EyeOff, LogIn, UserPlus } from 'lucide-react'
import axios from 'axios'
import { useNavigate } from 'react-router-dom'
import { useToast } from '../components/ToastProvider.jsx'

const initialForm = {
  username: '',
  password: '',
  re_password: '',
  invitation_code: ''
}

export default function LoginView() {
  const navigate = useNavigate()
  const toast = useToast()
  const [isLogin, setIsLogin] = useState(true)
  const [form, setForm] = useState(initialForm)
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const updateField = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }))
  }

  const toggleType = () => {
    setIsLogin((value) => !value)
    setForm(initialForm)
    setShowPassword(false)
  }

  const handleSubmit = async (event) => {
    event.preventDefault()

    if (!form.username || !form.password) {
      toast.warning('用户名和密码不能为空')
      return
    }

    if (!isLogin && form.password !== form.re_password) {
      toast.error('两次输入的密码不一致')
      return
    }

    if (!isLogin && !form.invitation_code) {
      toast.warning('请输入邀请码')
      return
    }

    setSubmitting(true)

    try {
      const url = isLogin ? '/api/v1/login' : '/api/v1/signup'
      const res = await axios.post(url, form)

      if (res.data.code === 1000) {
        if (isLogin) {
          const loginData = res.data.data
          localStorage.setItem('token', loginData.token)
          localStorage.setItem('username', loginData.user_name)
          localStorage.setItem('userid', loginData.user_id)
          toast.success('登录成功')
          navigate('/')
        } else {
          toast.success('注册成功，请登录')
          setIsLogin(true)
          setForm(initialForm)
        }
        return
      }

      toast.error(`失败：${res.data.msg}`)
    } catch (error) {
      console.error(error)
      toast.error('网络连接失败，请检查后端服务')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="login-container">
      <section className="auth-card">
        <div className="auth-header">
          <div className="auth-mark">GP</div>
          <h1>GamePulse 社区</h1>
          <p>{isLogin ? '账号登录' : '注册新账号'}</p>
        </div>

        <form className="auth-form" onSubmit={handleSubmit}>
          <label className="field-block">
            <span>用户名</span>
            <input
              value={form.username}
              placeholder="请输入用户名"
              autoComplete="username"
              onChange={(event) => updateField('username', event.target.value)}
            />
          </label>

          <label className="field-block">
            <span>密码</span>
            <div className="password-field">
              <input
                value={form.password}
                type={showPassword ? 'text' : 'password'}
                placeholder="请输入密码"
                autoComplete={isLogin ? 'current-password' : 'new-password'}
                onChange={(event) => updateField('password', event.target.value)}
              />
              <button
                className="icon-only password-toggle"
                type="button"
                aria-label={showPassword ? '隐藏密码' : '显示密码'}
                onClick={() => setShowPassword((value) => !value)}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </label>

          {!isLogin && (
            <>
              <label className="field-block">
                <span>确认密码</span>
                <input
                  value={form.re_password}
                  type={showPassword ? 'text' : 'password'}
                  placeholder="请再次输入密码"
                  autoComplete="new-password"
                  onChange={(event) => updateField('re_password', event.target.value)}
                />
              </label>

              <label className="field-block">
                <span>邀请码</span>
                <input
                  value={form.invitation_code}
                  type="text"
                  placeholder="请输入邀请码"
                  autoComplete="off"
                  inputMode="numeric"
                  onChange={(event) => updateField('invitation_code', event.target.value)}
                />
              </label>
            </>
          )}

          <button className="primary-action submit-btn" type="submit" disabled={submitting}>
            {isLogin ? <LogIn size={18} /> : <UserPlus size={18} />}
            <span>{submitting ? '提交中...' : isLogin ? '立即登录' : '注册'}</span>
          </button>
        </form>

        <button className="link-action auth-switch" type="button" onClick={toggleType}>
          {isLogin ? '没有账号？去注册' : '已有账号？去登录'}
        </button>
      </section>
    </main>
  )
}
