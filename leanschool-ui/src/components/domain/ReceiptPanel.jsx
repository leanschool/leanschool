import { useAuth } from '../../auth/useAuth'
import { hasFeature } from '../../auth/permissions'
import ReceiptList from '../ReceiptList'

export default function ReceiptPanel() {
  const { user } = useAuth()
  const canSubmit = hasFeature(user, 'submitReceipts')
  return <ReceiptList embedded canSubmit={canSubmit} />
}
