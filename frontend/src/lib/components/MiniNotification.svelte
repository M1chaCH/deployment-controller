<script lang="ts">
    import {createEventDispatcher} from 'svelte';

    export let message: string;

    const dispatch = createEventDispatcher()
</script>

<div class="notification">
    <span class="material-symbols-outlined">error</span>
    <p>{message}</p>
    <button class="icon-button" on:click={() => dispatch("close")} style="height: 2rem;">
        <span class="material-symbols-outlined">close</span>
    </button>
</div>

<style>
    .notification {
        width: 100%;
        background-color: var(--controller-area-color);
        border-left: 2px solid var(--michu-tech-warn);

        display: grid;
        grid-template-columns: 2rem calc(100% - 4rem) 2rem;
        align-items: center;

        padding: 0.5rem;
        margin: 0.5rem 0;
        box-sizing: border-box;

        animation: login-fly-in;
        animation-duration: 250ms;
        animation-timing-function: ease-out;
    }

    .notification span {
        color: var(--michu-tech-warn);
        font-size: 2rem;
    }
    
    .notification p {
        margin-left: 1rem;
    }

    @keyframes login-fly-in {
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