<script lang="ts">
    import {goto} from '$app/navigation';
    import {type AdminEditUserDto, type AdminPageDto, type AdminUserDeviceDto, type AdminUserDto, deletePage, deleteUser, getPages, getUsers, savePage, saveUser} from '$lib/api/admin';
    import {type ApiErrorDto, isErrorDto} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import AppCard from '$lib/components/AppCard.svelte';
    import ChangePassword from '$lib/components/ChangePassword.svelte';
    import FetchFailed from '$lib/components/FetchFailed.svelte';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import moment from 'moment';
    import {onMount} from 'svelte';
    import {v4 as uuidv4} from 'uuid';

    onMount(() => {
        userStore.subscribe(user => {
            if(user && (isErrorDto(user) || !user.admin)) {
                goto("/");
                return
            }
        })
    })

    let pageToDelete: string;
    let pageDeleteFailed: ApiErrorDto | undefined;

    let pageToEdit: AdminPageDto | undefined;
    let pageEditFailed: ApiErrorDto | undefined;
    let pageCreate: boolean = false;
    let openCreateEditPage = (create: boolean, page?: AdminPageDto) => {
        pageCreate = create;
        pageEditFailed = undefined;
        pageToEdit = {
            id: create ? uuidv4() : page!.id,
            privatePage: create ? true : page!.privatePage,
            description: create ? "" : page!.description,
            technicalName: create ? "" : page!.technicalName,
            title: create ? "" : page!.title,
            url: create ? "" : page!.url,
        }
    }

    let userToDelete: string;
    let userDeleteFailed: ApiErrorDto | undefined;

    let userResetPasswordId: string;
    let userResetPasswordMail: string;

    let userToEdit: AdminEditUserDto & { existingPages: Map<string, { access: boolean; name: string; }> } | undefined;
    let userEditFailed: ApiErrorDto | undefined;
    let userCreate: boolean = false;
    let openCreateEditUser = async (create: boolean, user?: AdminUserDto) => {
        userCreate = create;
        userEditFailed = undefined;
        userToEdit = {
            userId: create ? uuidv4() : user!.userId,
            mail: create ? "" : user!.mail,
            password: "",
            mfaType: create ? 'mfa-apptotp' : user!.mfaType,
            admin: create ? false : user!.admin,
            blocked: create ? false : user!.blocked,
            onboard: create ? false : user!.onboard,
            addPages: [],
            removePages: [],
            existingPages: create ? new Map<string, { access: boolean; name: string; }>()
                                  : new Map<string, { access: boolean; name: string; }>(user!.pageAccess
                                                                                             .filter(p => p.privatePage)
                                                                                             .map(p => [p.pageId, {
                                                                                                 access: p.hasAccess,
                                                                                                 name: p.technicalName,
                                                                                             }])),
        }
    }
    let togglePrivatePageAccess = (pageId: string) => {
        if(!userToEdit) {
            return
        }

        const current = userToEdit.existingPages.get(pageId);
        if(current!.access) {
            userToEdit.addPages.push(pageId)
            userToEdit.removePages = userToEdit.removePages.filter(p => p !== pageId)
        } else {
            userToEdit.removePages.push(pageId)
            userToEdit.addPages = userToEdit.addPages.filter(p => p !== pageId)
        }
    }

    let viewDevices: AdminUserDeviceDto[] | undefined;

    async function handleSavePage() {
        if(!pageToEdit) {
            return
        }

        const result = await savePage(pageToEdit, pageCreate)
        if(isErrorDto(result)) {
            pageEditFailed = result;
        } else {
            updateCaches();
        }
    }

    async function handleDeletePage() {
        if(pageToDelete) {
            const result = await deletePage(pageToDelete)
            if(isErrorDto(result)) {
                pageDeleteFailed = result
            } else {
                updateCaches();
            }
        }
    }

    async function handleSaveUser() {
        if(!userToEdit) {
            return
        }

        const result = await saveUser(userToEdit, userCreate)
        if(isErrorDto(result)) {
            userEditFailed = result;
        } else {
            updateCaches();
        }
    }

    async function handleDeleteUser() {
        if(userToDelete) {
            const result = await deleteUser(userToDelete)
            if(isErrorDto(result)) {
                userDeleteFailed = result
            } else {
                updateCaches();
            }
        }
    }

    function updateCaches() {
        location.reload(); // (;
    }
</script>

<PageOutline pageName="Administration">
    <div slot="description">
        <p>Keep track of users and manage page access.</p>
        <p class="subtext">Have fun!</p>
    </div>
    <div slot="content" class="quick-data">
        <p>Here will be some quick data once ready...</p>
    </div>
</PageOutline>
<div class="content">
    <h3>Pages</h3>
    <button class="carbon-button primary" on:click={() => openCreateEditPage(true)}>
        Create Page
        <span class="material-symbols-outlined icon">add</span>
    </button>
    <div class="cards-list">
        {#await getPages()}
            <p>Loading pages...</p>
        {:then pagesResponse}
            {#if isErrorDto(pagesResponse)}
                <FetchFailed error={pagesResponse} />
            {:else}
                {#each pagesResponse.sort((a, b) => {
                    if(a.privatePage && !b.privatePage) return -1
                    if(!a.privatePage && b.privatePage) return 1
                    return 0
                }) as page}
                    <AppCard title={page.title}>
                        <div slot="content">
                            <p><span class="subtext">Description:</span> {page.description}</p>
                            <p><span class="subtext">Technical name:</span> {page.technicalName}</p>
                            <p><span class="subtext">Url:</span> {page.url}</p>
                            <p><span class="subtext">Private:</span> {page.privatePage}</p>
                        </div>
                        <div slot="footer" class="controls">
                            <button class="carbon-button warn" on:click={() => pageToDelete = page.id}>
                                Delete
                                <span class="material-symbols-outlined icon">delete_forever</span>
                            </button>
                            <button class="carbon-button secondary" on:click={() => openCreateEditPage(false, page)}>
                                Edit
                                <span class="material-symbols-outlined icon">edit</span>
                            </button>
                        </div>
                    </AppCard>
                {/each}
            {/if}
        {/await}
    </div>

    <h3>Users</h3>
    <button class="carbon-button primary" on:click={() => openCreateEditUser(true)}>
        Create User
        <span class="material-symbols-outlined icon">add</span>
    </button>
    <div class="cards-list">
        {#await getUsers()}
            <p>Loading users...</p>
        {:then usersResponse}
            {#if isErrorDto(usersResponse)}
                <FetchFailed error={usersResponse} />
            {:else}
                {#each usersResponse as user}
                    <AppCard title={user.mail}>
                        <div slot="content">
                            <p><span class="subtext">Admin:</span> {user.admin}</p>
                            <p><span class="subtext">Onboard:</span> {user.onboard}</p>
                            <p><span class="subtext">Blocked:</span> {user.blocked}</p>
                            <p><span class="subtext">Last login:</span> {moment(user.lastLogin).format("DD.MM.yyyy HH:mm:ss")}</p>
                            <p><i>User was created at: {moment(user.createdAt).format("DD.MM.yyyy HH:mm:ss")}</i></p>
                            <div style="height: 1px; width: 75%; background-color: var(--controller-line-color); margin: 1rem 0;"></div>
                            {#each user.pageAccess.sort((a, b) => {
                                if(a.privatePage && !b.privatePage) return -1
                                if(!a.privatePage && b.privatePage) return 1
                                return 0
                            }) as page}
                                <p>
                                    <span class="subtext">
                                        {page.technicalName}
                                        {#if page.privatePage}
                                            <i>(Private)</i>
                                        {/if}:
                                    </span>{page.hasAccess ? 'Access granted' : 'Access denied'}
                                </p>
                            {/each}
                            <div style="height: 1px; width: 75%; background-color: var(--controller-line-color); margin: 1rem 0;"></div>
                            <button class="carbon-button flat" on:click={() => viewDevices = user.devices}>
                                View Devices
                                <span class="material-symbols-outlined">unfold_more</span>
                            </button>
                        </div>
                        <div slot="footer" class="controls">
                            <button class="carbon-button flat" on:click={() => {
                                userResetPasswordId = user.userId;
                                userResetPasswordMail = user.mail;
                            }}>
                                Change password
                                <span class="material-symbols-outlined icon">key</span>
                            </button>
                            <button class="carbon-button warn" on:click={() => userToDelete = user.userId}>
                                Delete
                                <span class="material-symbols-outlined icon">delete_forever</span>
                            </button>
                            <button class="carbon-button secondary" on:click={() => openCreateEditUser(false, user)}>
                                Edit
                                <span class="material-symbols-outlined icon">edit</span>
                            </button>
                        </div>
                    </AppCard>
                {/each}
            {/if}
        {/await}
    </div>
</div>

{#if pageToEdit}
    <div class="backdrop">
        <div class="content-card" style="max-height: 80vh; overflow-y: auto;">
            <form style="padding: 1rem 2rem;">
                <h3 style="margin-bottom: 1rem;">{pageCreate ? 'Create' : 'Edit'} page</h3>
                <div class="carbon-input">
                    <label for="pageTitle">Title</label>
                    <input id="pageTitle" type="text" bind:value={pageToEdit.title}>
                </div>
                <div class="carbon-input">
                    <label for="pageDesc">Description</label>
                    <input id="pageDesc" type="text" bind:value={pageToEdit.description}>
                </div>
                <div class="carbon-input">
                    <label for="pageTecName">Technical Name</label>
                    <input id="pageTecName" type="text" bind:value={pageToEdit.technicalName}>
                </div>
                <div class="carbon-input">
                    <label for="pageUrl">URL</label>
                    <input id="pageUrl" type="text" bind:value={pageToEdit.url}>
                </div>
                <div class="carbon-checkbox">
                    <input id="pagePrivate" type="checkbox" bind:checked={pageToEdit.privatePage}>
                    <label for="pagePrivate">Private</label>
                </div>
                {#if pageEditFailed}
                    <MiniNotification message={`(${pageEditFailed.status} - ${pageEditFailed.statusText}) - ${pageEditFailed.message}`} on:close={() => pageEditFailed = undefined} />
                {/if}
            </form>
            <div class="controls">
                <button class="carbon-button secondary" on:click={() => { pageToEdit = undefined; pageEditFailed = undefined; }}>
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Cancel
                </button>
                <button class="carbon-button warn" on:click={handleSavePage}>
                    {pageCreate ? 'Create' : 'Update'}
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
{/if}
{#if pageToDelete}
    <div class="backdrop">
        <div class="content-card">
            <div style="padding: 1rem 2rem;">
                <h3>Delete page</h3>
                <p>Do you really want to delete the page?</p>
                {#if pageDeleteFailed}
                    <MiniNotification message={`(${pageDeleteFailed.status} - ${pageDeleteFailed.statusText}) - ${pageDeleteFailed.message}`} on:close={() => pageDeleteFailed = undefined} />
                {/if}
            </div>
            <div class="controls">
                <button class="carbon-button secondary" on:click={() => { pageToDelete = ""; pageDeleteFailed = undefined; }}>
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Cancel
                </button>
                <button class="carbon-button warn" on:click={handleDeletePage}>
                    Proceed
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
{/if}

{#if userToEdit}
    <div class="backdrop">
        <div class="content-card" style="max-height: 80vh; overflow-y: auto;">
            <form style="padding: 1rem 2rem;">
                <h3 style="margin-bottom: 1rem;">{userCreate ? 'Create' : 'Edit'} user</h3>
                <div class="carbon-input">
                    <label for="userMail">Mail</label>
                    <input id="userMail" type="email" bind:value={userToEdit.mail}>
                </div>
                {#if userCreate}
                    <div class="carbon-input">
                        <label for="userPassword">Initial Password</label>
                        <input id="userPassword" type="text" bind:value={userToEdit.password}>
                    </div>
                {/if}
                <div class="carbon-radio-group">
                    <label>Choose MFA Type</label>
                    <div>
                        <input type="radio" id="mfa-type-app" name="mfa-type" value="mfa-apptotp" on:change={(e) => userToEdit.mfaType = e.target.value} checked={userToEdit.mfaType === 'mfa-apptotp'}/>
                        <label for="mfa-type-app">Authenticator App</label>
                    </div>
                    <div>
                        <input type="radio" id="mfa-type-mail" name="mfa-type" value="mfa-mailtotp" on:change={(e) => userToEdit.mfaType = e.target.value} checked={userToEdit.mfaType === 'mfa-mailtotp'}/>
                        <label for="mfa-type-mail">E-Mail</label>
                    </div>
                </div>
                {#if userToEdit.onboard}
                    <div class="carbon-checkbox">
                        <input id="userOnboard" type="checkbox" bind:checked={userToEdit.onboard}>
                        <label for="userOnboard">Onboard</label>
                    </div>
                {/if}
                <div class="carbon-checkbox">
                    <input id="userAdmin" type="checkbox" bind:checked={userToEdit.admin}>
                    <label for="userAdmin">Admin</label>
                </div>
                <div class="carbon-checkbox">
                    <input id="userBlocked" type="checkbox" bind:checked={userToEdit.blocked}>
                    <label for="userBlocked">Blocked</label>
                </div>
                {#if !userCreate}
                    <h4 style="padding: 0.5rem 0;">Page access</h4>
                    {#if !userToEdit.onboard}
                        <p class="subtext" style="padding-bottom: 0.5rem;">Changes will only be visible if user is onboard.</p>
                    {/if}
                {/if}
                {#each userToEdit.existingPages.entries() as [ pageId, page ]}
                    <div class="carbon-checkbox">
                        <input id={'id-' + pageId} type="checkbox" bind:checked={page.access} on:change={() => togglePrivatePageAccess(pageId)}>
                        <label for={'id-' + pageId}>{page.name}</label>
                    </div>
                {/each}

                {#if userEditFailed}
                    <MiniNotification message={`(${userEditFailed.status} - ${userEditFailed.statusText}) - ${userEditFailed.message}`} on:close={() => userEditFailed = undefined} />
                {/if}
            </form>
            <div class="controls">
                <button class="carbon-button secondary" on:click={() => { userToEdit = undefined; userEditFailed = undefined; }}>
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Cancel
                </button>
                <button class="carbon-button warn" on:click={handleSaveUser}>
                    {userCreate ? 'Create' : 'Update'}
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
{/if}
{#if userToDelete}
    <div class="backdrop">
        <div class="content-card">
            <div style="padding: 1rem 2rem;">
                <h3>Delete user</h3>
                <p>Do you really want to delete the user?</p>
                {#if userDeleteFailed}
                    <MiniNotification message={`(${userDeleteFailed.status} - ${userDeleteFailed.statusText}) - ${userDeleteFailed.message}`} on:close={() => userDeleteFailed = undefined} />
                {/if}
            </div>
            <div class="controls">
                <button class="carbon-button secondary" on:click={() => {userToDelete = ""; userDeleteFailed = undefined;}}>
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Cancel
                </button>
                <button class="carbon-button warn" on:click={handleDeleteUser}>
                    Proceed
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
{/if}
{#if userResetPasswordId && userResetPasswordMail}
    <ChangePassword userId={userResetPasswordId} targetOtherUserEmail={userResetPasswordMail} on:close={() => {
            userResetPasswordId = "";
            userResetPasswordMail = "";
        }}/>
{/if}
{#if viewDevices && viewDevices.length > 0}
    <div class="backdrop">
        <div class="content-card">
            <div style="padding: 1rem 2rem;">
                <h3>Devices</h3>
                {#each viewDevices as device}
                    <div style="height: 1px; width: 75%; background-color: var(--controller-line-color); margin: 1rem 0;"></div>
                    <div class="labeled-value">
                        <label for={`${device.deviceId}-ip`}>Ip Address</label>
                        <p id={`${device.deviceId}-ip`}>{device.ip}</p>
                    </div>
                    <div class="labeled-value">
                        <label for={`${device.deviceId}-agent`}>User Agent</label>
                        <p id={`${device.deviceId}-agent`} style="max-height: 4rem; overflow-y: auto;">{device.userAgent}</p>
                    </div>
                    <div class="labeled-value">
                        <label for={`${device.deviceId}-location`}>Location</label>
                        <p id={`${device.deviceId}-location`}>
                            {device.city}, {device.subdivision}, {device.country} - {device.systemOrganisation}
                        </p>
                    </div>
                    <div class="labeled-value">
                        <label for={`${device.deviceId}-clientId`}>Client ID</label>
                        <p id={`${device.deviceId}-clientId`}>{device.clientId}</p>
                    </div>
                    <div class="labeled-value">
                        <label for={`${device.deviceId}-deviceId`}>Device ID</label>
                        <p id={`${device.deviceId}-deviceId`}>{device.deviceId}</p>
                    </div>
                {/each}
            </div>
            <div class="controls">
                <button class="carbon-button secondary" on:click={() => viewDevices = undefined}>
                    <span class="material-symbols-outlined icon">arrow_left_alt</span>
                    Close
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .content {
        margin: 2rem;
    }

    .content h3 {
        margin-top: 2rem;
    }

    .cards-list {
        display: flex;
        flex-flow: row wrap;
        gap: 2rem;
    }

    .content-card {
        max-height: 80vh;
        overflow-y: auto;
    }

    .content-card .controls {
        position: sticky;
        bottom: 0;
    }

    .labeled-value {
        margin-top: 0.5rem;
        border-bottom: none;
        padding: 0;
    }
</style>