<?php

function startupError($message)
{
    $message = nl2br($message);
    $message = <<<HTML
<head>
<title>Faridoon startup error</title>
<style type = "text/css">
body {
    background-color: #efefef;
    font-family: sans-serif;
    padding: 2em;
}
</style>
</head>
<body>
    <h1>Faridoon startup error</h1>
    $message
</body>
HTML;
    echo $message;

    exit;
}

function requireDatabaseVersion(string $requiredMigration)
{
    try {
        $sql = 'SELECT id FROM migrations';
        $stmt = libAllure\DatabaseFactory::getInstance()->query($sql);
        $versionRows = array_column($stmt->fetchAll(), 'id');
    } catch (Exception $e) {
        startupError('Faridoon connected to the database, but the migrations table could not be queried.');
    }

    natsort($versionRows);
    $latestVersion = end($versionRows);

    if ($latestVersion != $requiredMigration) {
        if ($latestVersion == '') {
            $latestVersion = 'null';
        }

        startupError('Faridoon requires database version ' . $requiredMigration . ' but the database is at version ' . $latestVersion . '. Please <a href = "http://jamesread.github.io/Faridoon/installation/migrations/">run database migrations</a>.');
    }
}

function validateConfig() {
    // Validate ITEMS_PER_PAGE
    global $cfg;
    $limit = filter_var($cfg->get('ITEMS_PER_PAGE'), FILTER_VALIDATE_INT, ['options' => ['min_range' => 1]]);
    if ($limit === false) {
        startupError('ITEMS_PER_PAGE must be a positive integer');
    }
}
