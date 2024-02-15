"""Gmail.

Gmail is an email service provided by Google.

References:
  - API overview: https://developers.google.com/gmail/api/guides
  - REST API reference: https://developers.google.com/gmail/api/reference/rest
  - Go client API: https://pkg.go.dev/google.golang.org/api/gmail/v1
"""

def drafts_create(raw):
    pass

def drafts_delete(id):
    pass

def drafts_get(id, format):
    pass

def drafts_list(max_results, page_token, q, include_spam_trash):
    pass

def drafts_send(raw):
    pass

def drafts_update(id, raw):
    pass

def get_profile():
    pass

def history_list(start_history_id, max_results, page_token, label_id, history_types):
    pass

def labels_create(label):
    pass

def labels_delete(id):
    pass

def labels_get(id):
    pass

def labels_list():
    pass

def labels_patch(id, label):
    pass

def labels_update(id, label):
    pass

def messages_attachments_get(message_id, id):
    pass

def messages_batch_modify(ids, add_label_ids, remove_label_ids):
    pass

def messages_get(id, format, metadata_headers):
    pass

def messages_import(raw, internal_date_source, never_mark_spam, processForCalendar, deleted):
    pass

def messages_insert(raw, internal_date_source, deleted):
    pass

def messages_list(max_results, page_token, q, label_ids, include_spam_trash):
    pass

def messages_modify(id, add_label_ids, remove_label_ids):
    pass

def messages_send(raw):
    pass

def messages_trash(id):
    pass

def messages_untrash(id):
    pass

def threads_get(id, format, metadata_headers):
    pass

def threads_list(max_results, page_token, q, label_ids, include_spam_trash):
    pass

def threads_modify(id, add_label_ids, remove_label_ids):
    pass

def threads_trash(id):
    pass

def threads_untrash(id):
    pass