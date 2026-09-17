package test

import (
	"context"
	"testing"

	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type JoinCity struct {
	ID   uint
	Name string
}

type JoinProfile struct {
	ID     uint
	Bio    string
	CityID uint
	City   *JoinCity
}

type JoinUser struct {
	ID              uint
	Name            string
	MainProfileID   uint
	BackupProfileID uint
	MainProfile     *JoinProfile
	BackupProfile   *JoinProfile
	ManagerID       *uint
	Manager         *JoinUser
}

func TestContract_ExplicitAssociationAliases(t *testing.T) {
	d := openDB(t)
	require.NoError(t, d.AutoMigrate(&JoinCity{}, &JoinProfile{}, &JoinUser{}))
	city := &JoinCity{Name: "SF"}
	require.NoError(t, d.Create(city).Error)
	first := &JoinProfile{Bio: "first", CityID: city.ID}
	second := &JoinProfile{Bio: "second", CityID: city.ID}
	require.NoError(t, d.Create(first).Error)
	require.NoError(t, d.Create(second).Error)
	boss := &JoinUser{Name: "boss", MainProfileID: first.ID, BackupProfileID: second.ID}
	require.NoError(t, d.Create(boss).Error)
	require.NoError(t, d.Create(&JoinUser{Name: "worker", MainProfileID: first.ID, BackupProfileID: second.ID, ManagerID: &boss.ID}).Error)
	mainProfile := query.NewAssociation[JoinUser, JoinProfile]("MainProfile")
	backupProfile := query.NewAssociation[JoinUser, JoinProfile]("BackupProfile")
	cityRel := query.NewAssociation[JoinProfile, JoinCity]("City")
	manager := query.NewAssociation[JoinUser, JoinUser]("Manager")
	bio := query.NewStringColumn[string]("join_profiles", "bio")
	name := query.NewStringColumn[string]("join_users", "name")
	cityName := query.NewStringColumn[string]("join_cities", "name")
	read := func(t *testing.T, qb *query.Builder[JoinUser]) []JoinUser {
		t.Helper()
		var rows []JoinUser
		require.NoError(t, qb.Apply(d.WithContext(context.Background()).Model(&JoinUser{})).Find(&rows).Error)
		return rows
	}
	t.Run("two_relations_and_reused_column", func(t *testing.T) {
		condition := bio.Eq("first")
		qb := query.New[JoinUser]().
			Joins(mainProfile, bio.WithTable("MainProfile").Eq("first")).
			Joins(backupProfile, bio.WithTable("BackupProfile").Eq("second"))
		rows := read(t, qb)
		require.Len(t, rows, 2)
		require.Equal(t, "first", rows[0].MainProfile.Bio)
		require.Equal(t, "second", rows[0].BackupProfile.Bio)
		rows = read(t, query.New[JoinUser]().Preload(mainProfile, condition))
		require.Equal(t, "first", rows[0].MainProfile.Bio)
		var profiles []JoinProfile
		require.NoError(t, query.New[JoinProfile]().Where(condition).Apply(d.Model(&JoinProfile{})).Find(&profiles).Error)
		require.Len(t, profiles, 1)
	})
	t.Run("self_relation_and_explicit_parent", func(t *testing.T) {
		rows := read(t, query.New[JoinUser]().InnerJoins(manager, name.WithTable("Manager").Eq("boss"), name.Eq("worker")))
		require.Len(t, rows, 1)
		require.Equal(t, "worker", rows[0].Name)
		require.Equal(t, "boss", rows[0].Manager.Name)
		rows = read(t, query.New[JoinUser]().Joins(manager, name.WithTable("Manager").Eq("boss")).Where(name.Eq("worker")))
		require.Len(t, rows, 1)
		require.Equal(t, "worker", rows[0].Name)
	})
	t.Run("nested_relation_with_explicit_parent_join", func(t *testing.T) {
		rows := read(t, query.New[JoinUser]().
			Joins(mainProfile).
			Joins(mainProfile.Nested(cityRel), cityName.WithTable("MainProfile__City").Eq("SF")))
		require.Len(t, rows, 2)
		require.NotNil(t, rows[0].MainProfile)
		require.NotNil(t, rows[0].MainProfile.City)
		require.Equal(t, "SF", rows[0].MainProfile.City.Name)
	})
	t.Run("existing_parent_condition", func(t *testing.T) {
		rows := read(t, query.New[JoinUser]().
			Joins(mainProfile, bio.WithTable("MainProfile").Eq("missing")).
			Joins(mainProfile.Nested(cityRel), cityName.WithTable("MainProfile__City").Eq("SF")))
		require.Len(t, rows, 2)
		require.Nil(t, rows[0].MainProfile)
	})
	t.Run("nested_boolean_conditions", func(t *testing.T) {
		mainProfileBio := bio.WithTable("MainProfile")
		nested := func(tx *gorm.DB) *gorm.DB {
			return query.New[JoinProfile]().
				Where(mainProfileBio.In([]string{"first"})).
				Or(mainProfileBio.Eq("second")).
				Not(mainProfileBio.Eq("missing")).Apply(tx)
		}
		rows := read(t, query.New[JoinUser]().Where(name.Eq("worker")).InnerJoins(mainProfile, nested))
		require.Len(t, rows, 1)
		require.Equal(t, "first", rows[0].MainProfile.Bio)
	})
}
