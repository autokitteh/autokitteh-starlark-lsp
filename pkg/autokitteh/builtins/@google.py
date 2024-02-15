"""Google (All APIs).

Aggregation of all available Google APIs.

References:
  - REST API reference: https://developers.google.com/apis-explorer
  - Go client API: https://pkg.go.dev/google.golang.org/api
"""

def gmail_drafts_create(raw):
    pass

def gmail_drafts_delete(id):
    pass

def gmail_drafts_get(id, format):
    pass

def gmail_drafts_list(max_results, page_token, q, include_spam_trash):
    pass

def gmail_drafts_send(raw):
    pass

def gmail_drafts_update(id, raw):
    pass

def gmail_get_profile():
    pass

def gmail_history_list(start_history_id, max_results, page_token, label_id, history_types):
    pass

def gmail_labels_create(label):
    pass

def gmail_labels_delete(id):
    pass

def gmail_labels_get(id):
    pass

def gmail_labels_list():
    pass

def gmail_labels_patch(id, label):
    pass

def gmail_labels_update(id, label):
    pass

def gmail_messages_attachments_get(message_id, id):
    pass

def gmail_messages_batch_modify(ids, add_label_ids, remove_label_ids):
    pass

def gmail_messages_get(id, format, metadata_headers):
    pass

def gmail_messages_import(raw, internal_date_source, never_mark_spam, processForCalendar, deleted):
    pass

def gmail_messages_insert(raw, internal_date_source, deleted):
    pass

def gmail_messages_list(max_results, page_token, q, label_ids, include_spam_trash):
    pass

def gmail_messages_modify(id, add_label_ids, remove_label_ids):
    pass

def gmail_messages_send(raw):
    pass

def gmail_messages_trash(id):
    pass

def gmail_messages_untrash(id):
    pass

def gmail_threads_get(id, format, metadata_headers):
    pass

def gmail_threads_list(max_results, page_token, q, label_ids, include_spam_trash):
    pass

def gmail_threads_modify(id, add_label_ids, remove_label_ids):
    pass

def gmail_threads_trash(id):
    pass

def gmail_threads_untrash(id):
    pass

def sheets_a1_range(sheet_name: str|None, from: str|None, to: str|None):
    """https://developers.google.com/sheets/api/guides/concepts#expandable-1"""
    pass

def sheets_read_cell(spreadsheet_id: str, sheet_name: str|None, row_index: str, col_index: str, value_render_option: str|None):
    """Read a single cell"""
    pass

def sheets_read_range(spreadsheet_id: str, a1_range: str, value_render_option: str|None):
    """Read a range of cells"""
    pass

def sheets_set_background_color(spreadsheet_id: : str, a1_range: str, color: str):
    """Set the background color in a range of cells"""
    pass

def sheets_set_text_format(spreadsheet_id: str, a1_range: str, color: str|None, bold: str|None, italic: str|None, strikethrough: str|None, underline: str|None):
    """Set the text format in a range of cells"""
    pass

def sheets_write_cell(spreadsheet_id: str, sheet_name:str|None, row_index: str, col_index: str, value: str):
    """Write a single of cell"""
    pass

def sheets_write_range(spreadsheet_id: str, a1_range: str, data: str):
    """Write a range of cells"""
    pass
