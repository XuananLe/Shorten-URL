import http from 'k6/http'
import { check, sleep } from 'k6'

export default function () {
  const data = { username: 'username', password: 'password' }
  let res = http.post('https://httpbin.org/post', data)

  check(res, { 'success login': (r) => r.status === 200 })

  sleep(0.3)
}
