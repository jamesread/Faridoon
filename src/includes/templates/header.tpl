<!DOCTYPE html>

<html>
<head>
    <title>{$siteTitle}</title>

    <meta name="viewport" content="width=device-width, initial-scale=1">

    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <link rel = "stylesheet" type = "text/css" href = "resources/stylesheets/style.css" />
    <link rel = "stylesheet" type = "text/css" href = "resources/stylesheets/dark.css" />
    <link rel = "stylesheet" type = "text/css" href = "resources/stylesheets/app.css" />

	<link rel = "shortcut icon" type = "image/png" href = "resources/svg/faridoon.png" />

    <script src="resources/javascript/main.js" type = "text/javascript"></script>

	<style type = "text/css">
	{$inlineCss}
	</style>
</head>

<body>
    <header>
		<img src = "resources/svg/faridoon.svg" onclick = "javascript:clickLogo()" title = "logo" class = "logo" />

		<h1>
			<a href = "index.php">{$siteTitle}</a>
		</h1>

        <nav>
            <ul class = "navigation">
                <li><a href = "list.php?order=latest">Latest</a></li>
                <li><a href = "list.php?order=random">Random</a></li>
				{if $isVotingEnabled}
                <li><a href = "list.php?order=rank">Highest voted</a></li>
				{/if}
            </ul>
            <ul class = "navigation">
                {if $isLoggedIn}
                    <li><a href = "account.php"><strong>{$username}</strong></a></li>

					{if $hasApprovalPermissions}
						<li>
							<a href = "approvals.php">
								Approvals	({$countApprovals})
							</a>
						</li>
					{/if}
                {else}
					{if $isRegistrationEnabled}
						<li><a href = "login.php">Login</a></li>
					{/if}
                {/if}

				{if $isAddEnabled}
                <li class = "right"><a href = "add.php">Add</a></li>
				{/if}
            </ul>
			<ul class = "navigation" id = "developer-links" hidden>
				<li><a href = "#" onclick = "toggleFullscreen()">Fullscreen</a></li>
				<li><a href = "#" onclick = "document.getElementById('developer-links').hidden=true">&#10006;</a></li>
			</ul>
        </nav>

		<button id = "navigation-toggle" aria-label = "Open Sidebar Navigation" title = "Open Sidebar Navigation" aria-pressed = "false" aria-haspopup = "menu">&#9776;</button>

		<script type = "text/javascript">
		const button = document.getElementById("navigation-toggle");

		button.addEventListener("click", function() {
			var navigation = document.querySelector("nav");

			if (navigation.classList.contains('open')) {
				navigation.classList.remove('open');
				button.setAttribute("aria-pressed", "false");
				button.setAttribute("title", "Open Sidebar Navigation");
			} else {
				navigation.classList.add('open')
				button.setAttribute("aria-pressed", "true");
				button.setAttribute("title", "Close Sidebar Navigation");
			}
		});
		</script>
    </header>
    <main>
