import { useEffect, useState } from 'react';
import Form from 'react-bootstrap/Form';

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
    <Form>
      <Form.Check // prettier-ignore
        type="switch"
        id="custom-switch"
        label="Check this switch"
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
  );
}

export default BackgroundButton;