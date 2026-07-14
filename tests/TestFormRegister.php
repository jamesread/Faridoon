<?php

use PHPUnit\Framework\TestCase;
use libAllure\ElementAlphaNumeric;
use libAllure\Form;

class TestFormRegister extends TestCase
{
    protected function setUp(): void
    {
        Form::$registeredForms = [];
    }

    private function usernameElement(): ElementAlphaNumeric
    {
        $form = new faridoon\FormRegister();
        $username = $form->getElement('username');
        $this->assertInstanceOf(ElementAlphaNumeric::class, $username);

        return $username;
    }

    public function testUsernameElementIsAlphaNumeric()
    {
        $this->usernameElement();
    }

    public function testUsernameRejectsIllegalCharacters()
    {
        $username = $this->usernameElement();
        $username->setValue('user!');
        $username->validate();

        $this->assertNotNull($username->getValidationError());
    }

    public function testUsernameAcceptsAlphanumeric()
    {
        $username = $this->usernameElement();
        $username->setValue('gooduser');
        $username->validate();

        $this->assertNull($username->getValidationError());
    }

    public function testUsernameAcceptsSpacesAndUnderscores()
    {
        $username = $this->usernameElement();
        $username->setValue('good_user name');
        $username->validate();

        $this->assertNull($username->getValidationError());
    }
}
