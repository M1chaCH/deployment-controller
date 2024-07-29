import {PUBLIC_BACKEND_URL} from '$env/static/public';

export type LoginState = "logged-in" | "logged-out" | "two-factor-waiting" | "onboarding-waiting";

export interface UserInfoDto {
  userId: string,
  mail: string,
  admin: boolean,
  privatePages: string[],
  loginState: LoginState,
}

export async function postLogin(username: string, password: string): Promise<boolean> {
  return true;
}

export async function getUserInfo(): Promise<UserInfoDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/open/login`)
  const data = await response.json();

  if(response.ok) {
    return data;
  } else {
    const message = data?.message ?? response.statusText
    throw new Error(message)
  }
}