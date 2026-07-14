<?php

require_once '../vendor/autoload.php';

@include_once '/config/init.php';

use libAllure\ErrorHandler;

ErrorHandler::getInstance()->beGreedy();

$cfg = new \libAllure\ConfigFile();
$cfg->set(
    [
        'DB_NAME' => 'faridoon',
        'DB_HOST' => 'mysql',
        'SITE_TITLE' => 'Faridoon',
    ]
);
$cfg->loadFromPaths(
    [
        '/config/',
        '/var/www/html/faridoon/',
        '/etc/faridoon/config.ini',
    ]
);
$cfg->loadFromEnv();

use libAllure\Database;
use libAllure\DatabaseFactory;

$db = new Database($cfg->getDsn(), $cfg->get('DB_USER'), $cfg->get('DB_PASS'));
DatabaseFactory::registerInstance($db);

require_once 'includes/startup.php';

requireDatabaseVersion('3.charset-utf8mb4.sql');

require_once 'includes/functionality.php';

\libAllure\Sanitizer::getInstance()->enableSearchingPrefixKeys();

use libAllure\AuthBackend;
use libAllure\AuthBackendDatabase;

$backend = new AuthBackendDatabase($db);
$backend->register();

use libAllure\Session;

Session::setSessionName('faridoon');
Session::start();

use libAllure\Template;

$tpl = new Template(sys_get_temp_dir() . '/faridoon/' . 'includes/templates/');
$tpl->registerModifier('isAdmin', 'isAdmin');
