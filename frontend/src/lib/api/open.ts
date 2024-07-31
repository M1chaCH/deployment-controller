import {PUBLIC_BACKEND_URL} from '$env/static/public';
import type {LoginState} from '$lib/api/auth';

export interface ApiErrorDto {
  message: string;
  status: number;
  statusText: string;
}

export interface UserInfoDto {
  userId: string,
  mail: string,
  admin: boolean,
  onboard: boolean,
  loginState: LoginState,
}

export interface PageDto {
  pageTitle: string;
  pageDescription: string;
  pageUrl: string;
  privatePage: boolean;
  accessAllowed: boolean;
}

export function isErrorDto(obj: object | null | undefined): obj is ApiErrorDto {
  const dto = obj as ApiErrorDto
  if(!dto) return true;
  return dto.message !== undefined && dto.status !== undefined && dto.statusText !== undefined;
}

export async function getUserInfo(): Promise<UserInfoDto | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/open/login`, {
    credentials: 'include'
  })
  const data = await response.json();

  if(response.ok) {
    return data;
  } else {
    return {
      message: data?.message ?? response.statusText,
      status: response.status,
      statusText: response.statusText,
    }
  }
}

export async function getPages(): Promise<PageDto[] | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/open/pages`, {
    credentials: 'include'
  })
  const data = await response.json()

  if(response.ok) {
    return data;
  } else {
    return {
      message: data?.message ?? response.statusText,
      status: response.status,
      statusText: response.statusText,
    }
  }
}