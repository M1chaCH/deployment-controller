<script>
  export let apiUrl;

  let mail;
  let password;
  $: valid = !mail && !password;

  async function sendLogin() {
    const response = await fetch(`${apiUrl}/security/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ mail: mail, password: password }),
      credentials: 'include' // Include cookies in subsequent requests
    });

    if(response.ok) {
      console.log("successfully login");
      location.reload();
    } else
      console.warn("login failed", response)
  }
</script>

<form>
    <input type="text" placeholder="Mail" bind:value={mail} autocomplete="username"/>
    <input type="password" placeholder="Password" bind:value={password} autocomplete="current-password"/>
    <button type="submit" on:click|preventDefault={sendLogin} disabled={valid}>Login</button>
</form>