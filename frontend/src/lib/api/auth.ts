import {PUBLIC_BACKEND_URL} from '$env/static/public';

export type LoginState = "logged-in" | "logged-out" | "two-factor-waiting" | "onboarding-waiting";

interface LoginDto {
  mail: string;
  password: string;
}

interface MfaTokenDto {
  token: string;
}

export async function postLogin(username: string, password: string): Promise<LoginState> {
  const dto: LoginDto = {
    mail: username,
    password: password,
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
    return data.state ?? "logged-out"
  } else {
    return "logged-out"
  }
}

export async function postMfaValidation(mfaToken: string): Promise<boolean> {
  const dto: MfaTokenDto = {
    token: mfaToken,
  }
  const response = await fetch(`${PUBLIC_BACKEND_URL}/open/login/mfa`, {
    method: 'POST',
    body: JSON.stringify(dto),
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include'
  })

  return response.ok;
}

export async function postLogout(): Promise<void> {
  await fetch(`${PUBLIC_BACKEND_URL}/open/logout`, {
    method: 'POST',
    credentials: 'include'
  });
}