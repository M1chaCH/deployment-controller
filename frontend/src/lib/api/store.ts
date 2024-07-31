import {type ApiErrorDto, getPages, getUserInfo, isErrorDto, type PageDto, type UserInfoDto} from '$lib/api/open';
import {readable} from 'svelte/store';

export const userStore = readable<UserInfoDto | ApiErrorDto | null>(null, set => {
  // function is called to start the readable process, set can be called more than once
  getUserInfo().then(user => set(user)).catch(err => set({message: err.message, status: 0, statusText: "unknown"}))
  // function is used to cleanup, i don't have anything to cleanup
  return () => {}
})

export const pagesStore = readable<PageDto[] | ApiErrorDto | null>(null, set => {
  getPages().then(response => {
    if(isErrorDto(response)) {
      set(response)
    } else {
      set(response.sort((a, b) => {
        if(!a.accessAllowed && b.accessAllowed) return 1;
        if(a.accessAllowed && !b.accessAllowed) return -1;

        if(a.pageTitle < b.pageTitle) return -1;
        if(a.pageTitle > b.pageTitle) return 1;
        return 0;
      }))
    }
  }).catch(err => set({message: err.message, status: 0, statusText: "unknown"}))
  return () => {}
})