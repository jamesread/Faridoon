<?php

namespace faridoon;

use libAllure\ElementAlphaNumeric;
use libAllure\util\FormRegister as LibFormRegister;

class FormRegister extends LibFormRegister
{
    public function __construct()
    {
        parent::__construct();

        $this->addElement(new ElementAlphaNumeric('username', 'Username'));
    }
}
