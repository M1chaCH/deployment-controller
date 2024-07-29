import type {ApiErrorDto} from '$lib/api/api';
import {getUserInfo, type UserInfoDto} from '$lib/api/auth';
import {readable} from 'svelte/store';

export const userStore = readable<UserInfoDto | ApiErrorDto | null>(null, set => {
  // function is called to start the readable process, set can be called more than once
  getUserInfo().then(user => set(user)).catch(err => set({message: err.message}))
  // function is used to cleanup, i don't have anything to cleanup
  return () => {}
})