<script lang="ts">

    let mail: string;
    let password: string;

    $: console.log(mail, password);

    const login = () => {
      console.log("login")
    }
</script>

<h1>Login</h1>
<form>
    <input type="text" placeholder="Mail" bind:value={mail} autocomplete="username"/>
    <input type="password" placeholder="Password" bind:value={password} autocomplete="current-password"/>
    <button type="submit" on:click|preventDefault={login}>Login</button>
</form>
