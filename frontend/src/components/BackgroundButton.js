import { useEffect, useState } from 'react';
import Form from 'react-bootstrap/Form';
import './BackgroundButton.css';

function BackgroundButton() {
  const [theme, setTheme] = useState('table-wood');

  useEffect(() => {
    if (theme === 'table-wood') {
      document.body.className = 'theme-wood';
    } else {
      document.body.className = 'theme-green';
    }
  }, [theme]);

  return (
    <div className="background-button">
    <Form>
      <Form.Check // prettier-ignore
        type="switch"
        id="custom-switch"
        label="Toggle Background"
        checked={theme === 'table-wood'}
        onChange={() => setTheme(theme === 'table-wood' ? 'table-green' : 'table-wood')}
      />
      <Form // prettier-ignore
        disabled
        type="switch"
        label="disabled switch"
        id="disabled-custom-switch"
      />
    </Form>
    </div>
  );
}

export default BackgroundButton;