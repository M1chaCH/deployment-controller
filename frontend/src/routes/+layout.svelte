<script lang="ts">
    import {isErrorDto} from '$lib/api/open.js';
    import {userStore} from '$lib/api/store';
    import logo from '$lib/assets/michu-tech-icon-black-white.svg';
    import ThemeSwitcher from '$lib/colors/ThemeSwitcher.svelte';
    import ProfileImage from '$lib/ProfileImage.svelte';

    let username = "unknown"
    userStore.subscribe(user => {
        if(user && !isErrorDto(user)) {
            username = user.mail.split("@")[0] ?? "unknown"
        }
    })
</script>

<header>
    <a class="part title-part" href="/">
        <img alt="michu-tech logo small" src={logo} />
        <h1 class="title">MICHU - <span>TECH</span></h1>
    </a>
    <div class="part">
        <ThemeSwitcher />

        {#if $userStore !== null && !isErrorDto($userStore) && $userStore.loginState !== "logged-out" }
            <ProfileImage bind:username/>
        {:else }
            <a id="largeLoginButton" class="carbon-button secondary" href="/login#login">
                Login
                <span class="material-symbols-outlined icon">login</span>
            </a>
            <a id="smallLoginButton" class="icon-button" href="/login#login">
                <span class="material-symbols-outlined">login</span>
            </a>
        {/if}

    </div>
</header>
<main>
    <slot></slot>
</main>
<footer>
    <div class="footer-side me">
        <p id="footerName">Micha Schweizer</p>
        <div style="width: 100px; height: 1px; background-color: var(--controller-line-color); margin: 0.8rem;"></div>
        <p id="footerTitle">Application Developer</p>
    </div>
    <div class="footer-side links">
        <div class="links-container">
            <a href="mailto:admin@michu-tech.com">
                <span class="material-symbols-outlined">contact_mail</span>
                Contact
            </a>
            <a href="https://github.com/M1chaCH" target="_blank">
                <span class="material-symbols-outlined">code</span>
                Github
            </a>
            <a href="https://www.linkedin.com/in/micha-schweizer-83088a254/" target="_blank">
                <span class="material-symbols-outlined">domain</span>
                LinkedIn
            </a>
            <a href="https://www.instagram.com/ch_micha/" target="_blank">
                <span class="material-symbols-outlined">photo_camera</span>
                Instagram
            </a>
            <a href="/legal">
                <span class="material-symbols-outlined">policy</span>
                Legal & Privacy
            </a>
        </div>
    </div>
</footer>

<style>
    header {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;

        position: sticky;
        top: 0;
        z-index: 999;

        height: 4rem;
        box-sizing: border-box;

        background-color: var(--controller-area-color);
        outline: solid 1px var(--controller-line-color);

        animation: header-fly-in;
        animation-duration: 400ms;
        animation-timing-function: ease-out;
    }

    @keyframes header-fly-in {
        0% {
            transform: translateY(-100%);
        }
        100% {
            transform: translateY(0);
        }
    }

    header img {
        height: 100%;
        width: auto;
        aspect-ratio: 1/1;
        padding: 0.8rem;
        box-sizing: border-box;
    }

    header .part {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        height: 100%;
    }

    .title-part {
        padding-right: 1rem;
    }

    .title-part:hover {
        background-color: var(--controller-hover-color);
    }

    .title-part:focus, .title-part:active {
        background-color: var(--controller-focus-color);
    }

    .title {
        margin-left: 0.2rem;
        font-size: 1.8rem;
        font-weight: 200;
        letter-spacing: -0.05rem;
    }

    .title span {
        font-weight: 300;
    }

    #smallLoginButton {
        display: none;
    }

    @media (max-width: 550px) {
        #smallLoginButton {
            display: inherit;
        }

        #largeLoginButton {
            display: none;
        }
    }

    @media (max-width: 430px) {
        .title {
            font-size: 1.5rem;
            padding-right: 0;
            margin-left: 0.8rem;
        }
    }

    @media (max-width: 350px) {
        .title {
            display: none;
        }
    }

    footer {
        display: flex;
        flex-flow: row wrap;
        border-top: 1px solid var(--controller-line-color);
    }

    footer .footer-side {
        min-width: 220px;
        display: flex;
        flex-flow: column;
        justify-content: center;
        align-items: center;
        min-height: 50vh;
        padding: 1vh 5vw;
        box-sizing: border-box;
    }

    .me {
        flex: 1.5;
    }

    .links {
        flex: 1;
    }

    .footer-side a {
        display: flex;
        flex-flow: row nowrap;
        gap: 1rem;
        padding: 0.5rem;

        align-items: center;
    }

    .footer-side .links-container {
        border-left: solid 2px var(--michu-tech-accent);
        padding-left: 2rem;
    }

    #footerName {
        font-weight: 600;
        font-size: 3rem;
        text-align: center;
    }
</style>
