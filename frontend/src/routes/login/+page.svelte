<script lang="ts">
    import {type LoginState, postLogin} from '$lib/api/auth.js';
    import {isErrorDto} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import MiniNotification from '$lib/MiniNotification.svelte';
    import PageOutline from '$lib/PageOutline.svelte';
    import {onMount} from 'svelte';

    onMount(() => {
        // TODO user is never logged in -> cookie never sent to backend
        //  -> samesite looks at the two last domains (host.localhost & host.backend.localhost)
        // -> try "host.dev.localhost", "teachu.dev.localhost"... does this solve issue?
        userStore.subscribe(user => {
            if(user && !isErrorDto(user)) {
                mail = user.mail;

                if(user.loginState === "logged-in") {
                    location.href = parseTargetUrl()
                    return
                }

                if(user.loginState === "onboarding-waiting") {
                    location.href = "/onboarding"
                    return
                }
            }
        })
    })

    let mfa = false;
    let failed = false;
    $: formValid = mail?.length > 3 && password && (!mfa || (token?.length === 6 && !isNaN(parseInt(token))));

    let mail = "";
    let password = "";
    let token = "";

    async function login() {
        failed = false;
        mfa = false;

        if(!formValid) {
            failed = true;
        }

        let state: LoginState;
        try {
             state = await postLogin(mail, password);
        } catch (e) {
            console.error(e);
            state = "logged-out"
        }

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
</script>

<PageOutline pageName="Login">
    <div slot="description">
        <p>Logging in can grant you access to more services.</p>
        <p class="subtext">An account needs to be created by the admin.</p>
    </div>
    <div slot="content" class="page" id="login">
        <div class="login-card">
            <form class="login-inputs">
                <div class="carbon-input">
                    <label for="mail">E-Mail</label>
                    <input id="mail" type="email" bind:value={mail}/>
                </div>
                <div class="carbon-input">
                    <label for="password">Password</label>
                    <input id="password" type="password" bind:value={password}/>
                </div>
                {#if mfa}
                    <div class="carbon-input" id="tokenInput">
                        <label for="token">Token</label>
                        <input id="token" type="text" bind:value={token}/>
                        <p class="subtext">New Device, TwoFactor authentication required.</p>
                    </div>
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

    .login-card {
        border-left: 1px solid var(--controller-line-color);
        background-color: var(--controller-area-color);

        min-width: 220px;
        width: 80vw;
        max-width: 480px;
    }

    .login-inputs {
        padding: 1rem 2rem;
    }

    #tokenInput {
        animation: token-fly-in;
        animation-duration: 250ms;
        animation-timing-function: ease-out;
    }

    @keyframes token-fly-in {
        0% {
            opacity: 0;
            transform: translateY(-40%);
        }
        60% {
            opacity: 1;
        }
        100% {
            transform: translatey(0);
        }
    }
</style>