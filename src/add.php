<?php

require_once 'includes/widgets/header.php';

if (!isAddEnabled()) {
    simpleFatalError('Adding quotes is not available. Are you logged in?');
}

$f = new faridoon\FormQuote();

if ($f->validate()) {
    $f->process();

    $tpl->assign('quoteId', $db->lastInsertId());
    $tpl->display('quoteAdded.tpl');

    include_once 'includes/widgets/footer.php';
}

$tpl->displayForm($f);

require_once 'includes/widgets/footer.php';
