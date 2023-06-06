<script lang="ts">
  import {APP_CONFIG} from "../env";

  const pageUrl = APP_CONFIG.PAGE_URL;
  console.log(pageUrl)
  $: portfolioPage = `${pageUrl}/portfolio`
  $: automationPage = `${pageUrl}/automation`
</script>

<h1>Overview</h1>
<h3>Welcome to my overview, navigate to other pages</h3>
<a href="{portfolioPage}">Portfolio</a>
<a href={automationPage}>Automation</a>
<a href="/login">Login</a>