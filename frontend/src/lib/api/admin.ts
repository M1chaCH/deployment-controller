import {PUBLIC_BACKEND_URL} from '$env/static/public';
import type {ApiErrorDto, ApiSuccessDto, MfaType} from '$lib/api/open';

export interface AdminPageDto {
  id: string;
  technicalName: string;
  url: string;
  title: string;
  description: string;
  privatePage: boolean;
}

export interface AdminUserDto {
  userId: string
  mail: string
  admin: boolean
  blocked: boolean
  onboard: boolean
  createdAt: string
  lastLogin: string
  mfaType: MfaType
  pageAccess: AdminPageAccessDto[]
  devices: AdminUserDeviceDto[]
}

export interface AdminPageAccessDto {
  pageId: string
  technicalName: string
  hasAccess: boolean
  privatePage: boolean
}

export interface AdminUserDeviceDto {
  userId: string,
  deviceId: string,
  clientId: string,
  ip: string,
  userAgent: string,
  city: string,
  subdivision: string,
  country: string,
  systemOrganisation: string,
}

export interface AdminEditUserDto {
  userId: string;
  mail: string;
  password: string;
  mfaType: MfaType;
  admin: boolean;
  blocked: boolean;
  onboard: boolean;
  addPages: string[];
  removePages: string[];
}

export async function getPages(): Promise<AdminPageDto[] | ApiErrorDto> {
  try {
    const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/pages`, {
      credentials: 'include',
    })
    const data = await response.json()
    if(response.ok) {
      return data;
    }

    return {
      message: data.message ?? response.statusText,
      status: response.status,
      statusText: response.statusText,
    }
  } catch (e: any) {
    return {
      message: e?.message,
      status: 0,
      statusText: "unknown",
    }
  }
}

export async function savePage(dto: AdminPageDto, create: boolean): Promise<ApiSuccessDto | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/pages`, {
    credentials: 'include',
    method: create ? 'POST' : 'PUT',
    body: JSON.stringify(dto),
  })
  const data = await response.json()

  if(response.ok) {
    return data as ApiSuccessDto;
  }

  return {
    message: data.message ?? response.statusText,
    status: response.status,
    statusText: response.statusText,
  }
}

export async function deletePage(pageId: string): Promise<ApiSuccessDto | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/pages/${pageId}`, {
    credentials: 'include',
    method: 'DELETE',
  })
  const data = await response.json()

  if(response.ok) {
    return data as ApiSuccessDto;
  }

  return {
    message: data.message ?? response.statusText,
    status: response.status,
    statusText: response.statusText,
  }
}

export async function getUsers(): Promise<AdminUserDto[] | ApiErrorDto> {
  try {
    const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/users`, {
      credentials: 'include',
    })
    const data = await response.json()
    if(response.ok) {
      return data;
    }

    return {
      message: data.message ?? response.statusText,
      status: response.status,
      statusText: response.statusText,
    }
  } catch (e: any) {
    return {
      message: e?.message,
      statusText: "unknown",
      status: 0,
    }
  }
}

export async function saveUser(dto: AdminEditUserDto, create: boolean): Promise<ApiSuccessDto | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/users`, {
    credentials: 'include',
    method: create ? 'POST' : 'PUT',
    body: JSON.stringify(dto),
  })
  const data = await response.json()

  if(response.ok) {
    return data as ApiSuccessDto;
  }

  return {
    message: data.message ?? response.statusText,
    status: response.status,
    statusText: response.statusText,
  }
}

export async function deleteUser(userId: string): Promise<ApiSuccessDto | ApiErrorDto> {
  const response = await fetch(`${PUBLIC_BACKEND_URL}/admin/users/${userId}`, {
    credentials: 'include',
    method: 'DELETE',
  })
  const data = await response.json()

  if(response.ok) {
    return data as ApiSuccessDto;
  }

  return {
    message: data.message ?? response.statusText,
    status: response.status,
    statusText: response.statusText,
  }
}