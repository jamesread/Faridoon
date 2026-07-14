<?php

require_once 'includes/common.php';

if ($cfg->getBool('DISABLE_REGISTRATION')) {
    require_once 'includes/widgets/header.php';

    echo '<section>Registration is disabled.</section>';

    include_once 'includes/widgets/footer.php';
    exit;
}

use faridoon\FormRegister;

$f = new FormRegister();
$f->setTitle('Register as a new user');
$f->getElement('submit')->setCaption('Register user');

if ($f->validate()) {
    $f->process();

    $uid = $db->lastInsertId();

    $sql = 'SELECT * FROM users';
    $stmt = $db->prepare($sql);
    $stmt->execute();


    require_once 'includes/widgets/header.php';

    echo '<div class = "container">';
    echo '<section><p>You have been registered. You can now login.</p>';

    $sql = 'UPDATE users SET `group` = :gid WHERE id = :uid LIMIT 1';
    $stmt = $db->prepare($sql);
    $stmt->bindValue('uid', $uid);


    if ($stmt->numRows() == 1) {
        $gid = 1;

        echo '<p>You have been promoted to admin as you are the first registered user.</p>';
    } else {
        $gid = 2;
    }

    $stmt->bindValue('gid', $gid);
    $stmt->execute();

    echo '<p><a href = "login.php">Login</a></p>';
    echo '</section></div>';
} else {
    include_once 'includes/widgets/header.php';

    $tpl->displayForm($f);

    include_once 'includes/widgets/footer.php';
}
