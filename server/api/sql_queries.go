package api

// SQL query constants for server operations.

const listServersSQL = `
	SELECT
		s.id, s.name, s.host, s.ssh_port, s.ssh_user, s.server_type, s.carrier_type,
		COALESCE(s.monthly_rent, 0),
		s.status, s.last_report_at, s.agent_version, s.created_at,
		COALESCE(p.plan_type, 'no_limit'),
		COALESCE(p.total_quota, 0),
		COALESCE(p.used_bytes, 0),
		p.next_cycle_at, p.expire_at,
		p.cycle_day, p.cycle_time
	FROM servers s
	LEFT JOIN server_plans p ON p.server_id = s.id AND p.id = (
		SELECT id FROM server_plans WHERE server_id = s.id LIMIT 1
	)
	ORDER BY s.name
`

const getServerByIDSQL = `
	SELECT
		s.id, s.name, s.host, s.ssh_port, s.ssh_user, s.server_type, s.carrier_type,
		COALESCE(s.monthly_rent, 0),
		s.status, s.last_report_at, s.agent_version, s.created_at,
		COALESCE(p.plan_type, 'no_limit'),
		p.total_quota,
		COALESCE(p.used_bytes, 0),
		p.next_cycle_at, p.expire_at,
		p.cycle_day, p.cycle_time
	FROM servers s
	LEFT JOIN server_plans p ON p.server_id = s.id AND p.id = (
		SELECT id FROM server_plans WHERE server_id = s.id LIMIT 1
	)
	WHERE s.id = ?
`

const insertServerSQL = `
	INSERT INTO servers (name, host, ssh_port, ssh_user, server_type, carrier_type, monthly_rent, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, 'unknown', ?, ?)
`
