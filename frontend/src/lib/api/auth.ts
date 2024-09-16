import {PUBLIC_BACKEND_URL} from '$env/static/public';

export type LoginState = "logged-in" | "logged-out" | "two-factor-waiting" | "onboarding-waiting";

interface LoginDto {
  mail: string;
  password: string;
  token?: string;
}

export async function postLogin(username: string, password: string, token?: string): Promise<LoginState> {
  const dto: LoginDto = {
    mail: username,
    password: password,
    token: token,
  }
  const response = await fetch(`${PUBLIC_BACKEND_URL}/open/login`, {
    method: 'POST',
    body: JSON.stringify(dto),
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include'
  })

  if(response.ok) {
    const data = await response.json()
    return data.message ?? "logged-out"
  } else {
    return "logged-out"
  }
}