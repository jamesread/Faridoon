<section>
	<h2>Quote Added</h2>
	<p>Oh goodie. Another quote!</p>

	{if !isAdmin()}
        <p class = "good">Your quote needs approval before it shows up in the list.</p>
	{/if}

	<ul class = "block-links">
		<li><a href = "show.php?action=show&amp;id={$quoteId}">Show quote</a></li>
		<li><a href = "add.php">Add another</a></li>
	</ul>
</section>
