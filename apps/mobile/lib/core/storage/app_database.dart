import 'dart:io';
import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;

class SyncItem {
  final int id;
  final String endpoint;
  final String method;
  final String payloadJson;
  final DateTime createdAt;
  final bool isSynced;

  SyncItem({
    required this.id,
    required this.endpoint,
    required this.method,
    required this.payloadJson,
    required this.createdAt,
    required this.isSynced,
  });
}

class AppDatabase {
  final GeneratedDatabase db;

  AppDatabase([QueryExecutor? executor])
      : db = _SimpleDb(executor ?? _openConnection());

  Future<void> initSchema() async {
    await db.customStatement('''
      CREATE TABLE IF NOT EXISTS sync_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        endpoint TEXT NOT NULL,
        method TEXT NOT NULL,
        payload_json TEXT NOT NULL,
        created_at TEXT NOT NULL,
        is_synced INTEGER NOT NULL DEFAULT 0
      )
    ''');
  }

  Future<int> enqueueSyncItem(String endpoint, String method, String payloadJson) async {
    await initSchema();
    final result = await db.customInsert(
      'INSERT INTO sync_queue (endpoint, method, payload_json, created_at, is_synced) VALUES (?, ?, ?, ?, 0)',
      variables: [
        Variable.withString(endpoint),
        Variable.withString(method),
        Variable.withString(payloadJson),
        Variable.withString(DateTime.now().toIso8601String()),
      ],
    );
    return result;
  }

  Future<List<SyncItem>> getPendingSyncItems() async {
    await initSchema();
    final rows = await db.customSelect(
      'SELECT id, endpoint, method, payload_json, created_at, is_synced FROM sync_queue WHERE is_synced = 0',
    ).get();

    return rows.map((row) {
      return SyncItem(
        id: row.read<int>('id'),
        endpoint: row.read<String>('endpoint'),
        method: row.read<String>('method'),
        payloadJson: row.read<String>('payload_json'),
        createdAt: DateTime.parse(row.read<String>('created_at')),
        isSynced: row.read<int>('is_synced') == 1,
      );
    }).toList();
  }

  Future<void> markSynced(int id) async {
    await initSchema();
    await db.customUpdate(
      'UPDATE sync_queue SET is_synced = 1 WHERE id = ?',
      variables: [Variable.withInt(id)],
    );
  }
}

class _SimpleDb extends GeneratedDatabase {
  _SimpleDb(QueryExecutor e) : super(e);
  @override
  int get schemaVersion => 1;
  @override
  List<TableInfo> get allTables => [];
}

LazyDatabase _openConnection() {
  return LazyDatabase(() async {
    final dbFolder = await getApplicationDocumentsDirectory();
    final file = File(p.join(dbFolder.path, 'rt_digital.sqlite'));
    return NativeDatabase(file);
  });
}

