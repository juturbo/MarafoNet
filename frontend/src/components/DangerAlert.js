import Alert from 'react-bootstrap/Alert';
import Button from 'react-bootstrap/Button';
import './TableScreen.css';

function DangerAlert({ message, onClose }) {
  console.log('DangerAlert received message:', message);
  if (!message) return null;

  return (
    <div className="danger-alert-overlay">
      <Alert variant="danger" dismissible onClose={onClose}>
        <Alert.Heading>Error</Alert.Heading>
        <p>{message}</p>
        <div className="d-flex justify-content-end">
          <Button onClick={onClose} variant="outline-danger" size="sm">
            Dismiss
          </Button>
        </div>
      </Alert>
    </div>
  );
}

export default DangerAlert;