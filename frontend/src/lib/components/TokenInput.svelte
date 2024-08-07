<script lang="ts">

    import {createEventDispatcher, onMount} from 'svelte';

    let inputEl1: HTMLInputElement;
    let inputEl2: HTMLInputElement;
    let inputEl3: HTMLInputElement;
    let inputEl4: HTMLInputElement;
    let inputEl5: HTMLInputElement;
    let inputEl6: HTMLInputElement;

    const dispatch = createEventDispatcher();

    onMount(() => {
        inputEl1 = document.getElementById("digitInput1") as HTMLInputElement;
        inputEl2 = document.getElementById("digitInput2") as HTMLInputElement;
        inputEl3 = document.getElementById("digitInput3") as HTMLInputElement;
        inputEl4 = document.getElementById("digitInput4") as HTMLInputElement;
        inputEl5 = document.getElementById("digitInput5") as HTMLInputElement;
        inputEl6 = document.getElementById("digitInput6") as HTMLInputElement;
    })

    function updateValue() {
        const value = inputEl1.value + inputEl2.value + inputEl3.value + inputEl4.value + inputEl5.value + inputEl6.value;
        dispatch('input', { value: value });
    }

    function inputChanged(event: KeyboardEvent, previousElement?: HTMLInputElement, nextElement?: HTMLInputElement) {
        if(event.altKey || event.ctrlKey || event.metaKey || (event.key.length > 1 && event.key !== "Backspace")) {
            return
        }

        event.preventDefault()

        const currentElement = event.target as HTMLInputElement;

        if(event.key === "Backspace") {
            if(currentElement.value === "" && previousElement) {
                previousElement.value = "";
                previousElement.focus();
            } else {
                currentElement.value = "";
            }
        } else {
            currentElement.value = event.key;

            if (nextElement) {
                nextElement.focus();
            }
        }
        updateValue()
    }

    function handlePaste(event: ClipboardEvent) {
        if(!event.clipboardData) return

        const data = event.clipboardData.getData("text").split("")
        inputEl1.value = data[0] ?? "";
        inputEl2.value = data[1] ?? "";
        inputEl3.value = data[2] ?? "";
        inputEl4.value = data[3] ?? "";
        inputEl5.value = data[4] ?? "";
        inputEl6.value = data[5] ?? "";
        updateValue()
    }
</script>

<p class="label">Token</p>
<form class="token-input">
    <input id="digitInput1" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, undefined, inputEl2)}/>
    <input id="digitInput2" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, inputEl1, inputEl3)}/>
    <input id="digitInput3" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, inputEl2, inputEl4)}/>
    <input id="digitInput4" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, inputEl3, inputEl5)}/>
    <input id="digitInput5" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, inputEl4, inputEl6)}/>
    <input id="digitInput6" class="digit-input" maxlength="1" on:paste|preventDefault={(e) => handlePaste(e)} on:keydown={(e) => inputChanged(e, inputEl5)}/>
</form>

<style>
    .label {
        font-size: 0.8rem;
        font-weight: 200;
        margin-bottom: 0.1rem;
    }

    .token-input {
        display: flex;
        flex-flow: row nowrap;
        gap: 0.5rem;

        align-items: center;
        justify-content: center;
    }

    .digit-input {
        background-color: transparent;
        border: 1px solid var(--controller-line-color);
        outline: none;
        font-size: 1.4rem;

        width: 1.5rem;
        height: 1.5rem;
        padding: 0.5rem;
        text-align: center;

        font-family: "IBM Plex Sans", "Arial", sans-serif;
        color: var(--michu-tech-foreground);
        text-transform: uppercase;

        transition: all 120ms ease-out;
    }

    .digit-input:hover {
        background-color: var(--controller-hover-color);
    }

    .digit-input:focus, .digit-input:active {
        background-color: var(--controller-focus-color);
    }
</style>