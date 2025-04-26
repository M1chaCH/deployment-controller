<script lang="ts">
    import {goto} from '$app/navigation';
    import {type LoginState, postLogin, postMfaValidation} from '$lib/api/auth.js';
    import {isErrorDto, type MfaType, postSendMfaMail} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/components/MiniNotification.svelte';
    import PageOutline from '$lib/components/PageOutline.svelte';
    import TokenInput from '$lib/components/TokenInput.svelte';
    import {onMount} from 'svelte';

    onMount(() => {
        userStore.subscribe(user => {
            if(user && !isErrorDto(user)) {
                mail = user.mail;
                mfaType = user.mfaType;

                if(user.loginState === "logged-in") {
                    goto(parseTargetUrl())
                    return
                }

                if(user.loginState === "onboarding-waiting") {
                    goto("/onboarding");
                    return
                }

                mfa = user.loginState === "two-factor-waiting";
            }
        })

        // e-mail is sent automatically, let user wait initially
        sendingMail = true;
        setTimeout(() => sendingMail = false, 20 * 1000)
    })

    let mfa = false;
    let mfaType: MfaType = 'mfa-apptotp';
    let failed = false;
    $: formValid = (!mfa && mail?.length > 3 && password) || (mfa && token?.length === 6 && !isNaN(parseInt(token)));
    let sendingMail = true;
    let sendMfaMailFailed = false;

    let mail = "";
    let password = "";
    let token = "";

    async function login() {
        failed = false;
        if(!formValid) {
            failed = true;
            return
        }

        let state: LoginState;
        try {
            if(mfa) {
                const tokenValid = await postMfaValidation(token);
                if(tokenValid) {
                    state = "logged-in";
                } else {
                    failed = true;
                    state = "two-factor-waiting"
                }
            } else {
                state = await postLogin(mail, password);
            }
        } catch (e) {
            console.error(e);
            state = "logged-out"
        }

        // these redirects can't be routed using the svelte kit router because I'd like to reload the user, which is in a readable store.
        switch (state) {
            case 'logged-in':
                location.href = parseTargetUrl()
                return
            case 'onboarding-waiting':
                location.href = "/onboarding"
                return
            case 'logged-out':
                failed = true
                break
            case 'two-factor-waiting':
                mfa = true
                break
        }
    }

    function parseTargetUrl() {
        const props = new URLSearchParams(window.location.search);
        const origin = props.get("origin");

        if(origin) {
            return `https://${origin}`;
        }

        return "/";
    }

    async function sendMfaMail() {
        sendingMail = true;
        sendMfaMailFailed = false;

        const response = await postSendMfaMail()
        if(isErrorDto(response)) {
            sendMfaMailFailed = true;
        }
        sendingMail = false;
    }

    function handleInputKeydown(e: KeyboardEvent) {
        if(e.key === "Enter") {
            e.preventDefault();
            login();
        }
    }
</script>

<svelte:head>
    <title>Micha Schweizer @ Login</title>
</svelte:head>

<PageOutline pageName="Login">
    <div slot="description">
        <p>Logging in can grant you access to more services.</p>
        <p class="subtext">An account needs to be created by the admin.</p>
    </div>
    <div slot="content" class="page" id="login">
        <div class="content-card">
            <form class="login-inputs" on:submit|preventDefault={login}>
                {#if !mfa}
                    <div class="carbon-input">
                        <label for="mail">E-Mail</label>
                        <input id="mail" type="email" bind:value={mail}/>
                    </div>
                    <div class="carbon-input">
                        <label for="password">Password</label>
                        <input id="password" type="password" bind:value={password} on:keydown={handleInputKeydown}/>
                    </div>
                {:else}
                    {#if mfaType === 'mfa-mailtotp'}
                        <button style="margin: 8px;" class="carbon-button primary" on:click={() => sendMfaMail()} disabled={sendingMail}>Send E-Mail again</button>
                        {#if sendMfaMailFailed}
                            <MiniNotification message="Failed to send MFA Token via mail." on:close={() => sendMfaMailFailed = false} />
                        {/if}
                    {/if}

                    <TokenInput on:input={e => token = e.detail.value}/>
                    <p class="subtext" style="margin-top: 12px;">New Device, TwoFactor authentication required.</p>
                {/if}
            </form>
            {#if failed}
                <MiniNotification message="Login failed!" on:close={() => failed = false} />
            {/if}
            <div class="controls">
                <a class="carbon-button flat" href="mailto:admin@michu-tech.com?subject=Request Account">
                    Request Account
                </a>
                <button class="carbon-button primary" on:click|preventDefault={login} disabled={!formValid}>
                    Login
                    <span class="material-symbols-outlined icon">arrow_right_alt</span>
                </button>
            </div>
        </div>
    </div>
</PageOutline>

<style>
    .page {
        display: flex;
        flex-flow: column;
        /* page - header height */
        min-height: calc(100vh - 4rem);
        align-items: center;
        justify-content: center;
    }

    .login-inputs {
        padding: 1rem 2rem;
    }
</style>